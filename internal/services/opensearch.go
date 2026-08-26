package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const maximumOpenSearchLock = 16_384

var (
	openSearchLineRE       = regexp.MustCompile(`^([a-z_]+)='([^']+)'$`)
	openSearchRepositoryRE = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]*$`)
	openSearchVersionRE    = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	openSearchDigestRE     = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

// OpenSearchImages is a repository-owned, digest-pinned compatibility pair.
type OpenSearchImages struct {
	Repository string
	OldVersion string
	OldDigest  string
	NewVersion string
	NewDigest  string
}

// ParseOpenSearchImages parses the supported declarative image-lock format
// without evaluating it as shell input.
func ParseOpenSearchImages(reader io.Reader) (OpenSearchImages, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maximumOpenSearchLock+1))
	if err != nil {
		return OpenSearchImages{}, fmt.Errorf("read OpenSearch image lock: %w", err)
	}
	if len(data) > maximumOpenSearchLock {
		return OpenSearchImages{}, errors.New("OpenSearch image lock exceeds limit")
	}
	values := make(map[string]string, 5)
	content := strings.TrimSuffix(string(data), "\n")
	for _, line := range strings.Split(content, "\n") {
		matches := openSearchLineRE.FindStringSubmatch(line)
		if len(matches) != 3 {
			return OpenSearchImages{}, errors.New("OpenSearch image lock contains a malformed line")
		}
		if _, duplicate := values[matches[1]]; duplicate {
			return OpenSearchImages{}, fmt.Errorf("OpenSearch image lock repeats %s", matches[1])
		}
		values[matches[1]] = matches[2]
	}
	allowed := map[string]bool{
		"opensearch_image_repository": true, "opensearch_old_version": true,
		"opensearch_old_digest": true, "opensearch_new_version": true,
		"opensearch_new_digest": true,
	}
	for key := range values {
		if !allowed[key] {
			return OpenSearchImages{}, fmt.Errorf("OpenSearch image lock contains unknown key %s", key)
		}
	}
	images := OpenSearchImages{
		Repository: values["opensearch_image_repository"], OldVersion: values["opensearch_old_version"],
		OldDigest: values["opensearch_old_digest"], NewVersion: values["opensearch_new_version"],
		NewDigest: values["opensearch_new_digest"],
	}
	if err := images.validate(); err != nil {
		return OpenSearchImages{}, err
	}
	return images, nil
}

func (images OpenSearchImages) validate() error {
	if !openSearchRepositoryRE.MatchString(images.Repository) || strings.Contains(images.Repository, "..") || strings.Contains(images.Repository, "//") {
		return errors.New("OpenSearch image policy repository is malformed")
	}
	if !openSearchVersionRE.MatchString(images.OldVersion) || !openSearchVersionRE.MatchString(images.NewVersion) {
		return errors.New("OpenSearch image policy version is malformed")
	}
	if !openSearchDigestRE.MatchString(images.OldDigest) || !openSearchDigestRE.MatchString(images.NewDigest) {
		return errors.New("OpenSearch image policy digest is malformed")
	}
	if images.OldVersion == images.NewVersion || images.OldDigest == images.NewDigest {
		return errors.New("OpenSearch image policy versions and digests must differ")
	}
	return nil
}

// OldImage returns the immutable old compatibility image.
func (images OpenSearchImages) OldImage() string { return images.Repository + "@" + images.OldDigest }

// NewImage returns the immutable current compatibility image.
func (images OpenSearchImages) NewImage() string { return images.Repository + "@" + images.NewDigest }

func (images OpenSearchImages) definition() (definition, error) {
	if err := images.validate(); err != nil {
		return definition{}, err
	}
	return definition{
		name: "opensearch", image: images.NewImage(), port: 9200, requiresHTTP: true,
		runArguments: func(string) []string {
			return []string{
				"--cpus=1", "--memory=1g", "--pids-limit=512", "--ulimit", "nofile=1024:1024",
				"-e", "discovery.type=single-node", "-e", "DISABLE_SECURITY_PLUGIN=true",
				"-e", "OPENSEARCH_JAVA_OPTS=-Xms512m -Xmx512m",
			}
		},
		environment: func(port string) map[string]string {
			return map[string]string{
				"OPENSEARCH_URL":              "http://127.0.0.1:" + port,
				"OPENSEARCH_EXPECTED_VERSION": images.NewVersion,
			}
		},
	}, nil
}

func httpProbe(ctx context.Context, url string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("redirects are not allowed")
		},
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return validateHTTPStatus(response.StatusCode)
}

func validateHTTPStatus(status int) error {
	if status < http.StatusOK {
		return fmt.Errorf("OpenSearch readiness returned HTTP %d", status)
	}
	if status >= http.StatusMultipleChoices {
		return fmt.Errorf("OpenSearch readiness returned HTTP %d", status)
	}
	return nil
}
