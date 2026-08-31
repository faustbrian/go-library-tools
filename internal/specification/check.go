// Package specification validates specification decision registers and their
// executable conformance evidence.
package specification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

const (
	maximumDocumentSize   = 8 << 20
	maximumEvidenceSize   = 8 << 20
	maximumAuthoritySize  = 16 << 20
	maximumAuthorities    = 64
	maximumOnlineDuration = 2 * time.Minute
)

var (
	decisionHeading           = regexp.MustCompile(`(?m)^## ([A-Z][A-Z0-9-]*-DEC-[0-9]{3}): ([^\n]+)$`)
	levelTwoHeading           = regexp.MustCompile(`(?m)^## [^\n]+$`)
	decisionID                = regexp.MustCompile(`\b[A-Z][A-Z0-9-]*-DEC-[0-9]{3}\b`)
	markdownLink              = regexp.MustCompile(`\]\([[:space:]]*(?:<([^>\n]+)>|([^[:space:])]+))(?:[[:space:]]+(?:"[^"\n]*"|'[^'\n]*'|\([^\)\n]*\)))?[[:space:]]*\)`)
	sha256Value               = regexp.MustCompile(`^[0-9a-f]{64}$`)
	publicAuthorityExceptions = []netip.Prefix{
		netip.MustParsePrefix("192.0.0.9/32"),
		netip.MustParsePrefix("192.0.0.10/32"),
		netip.MustParsePrefix("2001:1::1/128"),
		netip.MustParsePrefix("2001:1::2/128"),
		netip.MustParsePrefix("2001:1::3/128"),
		netip.MustParsePrefix("2001:3::/32"),
		netip.MustParsePrefix("2001:4:112::/48"),
		netip.MustParsePrefix("2001:20::/28"),
		netip.MustParsePrefix("2001:30::/28"),
	}
	nonPublicAuthorityPrefixes = []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),
		netip.MustParsePrefix("100.64.0.0/10"),
		netip.MustParsePrefix("192.0.0.0/24"),
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("192.88.99.0/24"),
		netip.MustParsePrefix("198.18.0.0/15"),
		netip.MustParsePrefix("198.51.100.0/24"),
		netip.MustParsePrefix("203.0.113.0/24"),
		netip.MustParsePrefix("240.0.0.0/4"),
		netip.MustParsePrefix("64:ff9b:1::/48"),
		netip.MustParsePrefix("100::/64"),
		netip.MustParsePrefix("100:0:0:1::/64"),
		netip.MustParsePrefix("2001::/23"),
		netip.MustParsePrefix("2001:db8::/32"),
		netip.MustParsePrefix("2002::/16"),
		netip.MustParsePrefix("3fff::/20"),
		netip.MustParsePrefix("5f00::/16"),
		netip.MustParsePrefix("fec0::/10"),
	}
)

// Report summarizes the specification-backed modules and registered choices.
type Report struct {
	Modules    int
	Decisions  int
	Unresolved int
}

// CheckOnline validates the static register contract and verifies that mutable
// authoritative errata and release feeds still match their reviewed digests.
func CheckOnline(ctx context.Context, root string, catalog inventory.Inventory, client *http.Client) (Report, error) {
	return checkOnline(ctx, root, catalog, client, net.DefaultResolver)
}

type authorityResolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

func checkOnline(ctx context.Context, root string, catalog inventory.Inventory, client *http.Client, resolver authorityResolver) (Report, error) {
	report, policies, err := check(root, catalog)
	if err != nil {
		return report, err
	}
	if report.Modules == 0 {
		return report, err
	}
	if report.Unresolved > 0 {
		return report, fmt.Errorf("specification decisions require review: %d unresolved", report.Unresolved)
	}
	if client == nil {
		return Report{}, errors.New("specification authority HTTP client is required")
	}
	client, err = secureAuthorityClient(client, resolver)
	if err != nil {
		return Report{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, maximumOnlineDuration)
	defer cancel()
	for _, policy := range policies {
		for _, item := range policy.Authorities {
			if err := fetchAuthorityResolved(ctx, client, resolver, item); err != nil {
				return Report{}, err
			}
		}
	}
	return report, nil
}

func fetchAuthority(ctx context.Context, client *http.Client, item authority) error {
	return fetchAuthorityResolved(ctx, client, net.DefaultResolver, item)
}

func fetchAuthorityResolved(ctx context.Context, client *http.Client, resolver authorityResolver, item authority) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, item.URL, nil)
	if err != nil {
		return fmt.Errorf("create specification authority request %q: %w", item.ID, err)
	}
	boundedClient := *client
	boundedClient.CheckRedirect = func(candidate *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		if candidate.URL.Scheme != "https" || !strings.EqualFold(candidate.URL.Host, request.URL.Host) {
			return errors.New("redirect escaped its pinned HTTPS authority")
		}
		return validateAuthorityDestination(candidate.Context(), resolver, candidate.URL)
	}
	if err := validateAuthorityDestination(ctx, resolver, request.URL); err != nil {
		return fmt.Errorf("fetch specification authority %q: %w", item.ID, err)
	}
	response, err := boundedClient.Do(request)
	if err != nil {
		return fmt.Errorf("fetch specification authority %q: %w", item.ID, err)
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, maximumAuthoritySize+1))
	closeErr := response.Body.Close()
	if readErr != nil {
		return fmt.Errorf("read specification authority %q: %w", item.ID, readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close specification authority %q: %w", item.ID, closeErr)
	}
	if len(body) > maximumAuthoritySize {
		return fmt.Errorf("specification authority %q exceeds maximum size", item.ID)
	}
	if item.Access == "restricted" {
		if response.StatusCode != item.ExpectedStatus {
			return fmt.Errorf("restricted specification authority %q returned HTTP %d, want %d", item.ID, response.StatusCode, item.ExpectedStatus)
		}
		return nil
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("specification authority %q returned HTTP %d", item.ID, response.StatusCode)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(body))
	if digest != item.SHA256 {
		return fmt.Errorf("specification authority %q changed and requires review", item.ID)
	}
	return nil
}

func validateAuthorityDestination(ctx context.Context, resolver authorityResolver, destination *url.URL) error {
	_, err := resolveAuthorityDestination(ctx, resolver, destination)
	return err
}

func resolveAuthorityDestination(ctx context.Context, resolver authorityResolver, destination *url.URL) ([]netip.Addr, error) {
	if resolver == nil || destination == nil || destination.Hostname() == "" {
		return nil, errors.New("authority resolver and destination are required")
	}
	addresses, err := resolver.LookupNetIP(ctx, "ip", destination.Hostname())
	if err != nil {
		return nil, fmt.Errorf("resolve authority destination: %w", err)
	}
	if len(addresses) == 0 {
		return nil, errors.New("authority destination resolved no addresses")
	}
	for index, address := range addresses {
		address = address.Unmap()
		if !isPublicAuthorityAddress(address) {
			return nil, fmt.Errorf("authority must resolve only to a public network destination: %s", address)
		}
		addresses[index] = address
	}
	return addresses, nil
}

func isPublicAuthorityAddress(address netip.Addr) bool {
	if !address.IsValid() || !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.Zone() != "" {
		return false
	}
	if slices.ContainsFunc(publicAuthorityExceptions, func(prefix netip.Prefix) bool {
		return prefix.Contains(address)
	}) {
		return true
	}
	return !slices.ContainsFunc(nonPublicAuthorityPrefixes, func(prefix netip.Prefix) bool {
		return prefix.Contains(address)
	})
}

type authorityDialer func(context.Context, string, string) (net.Conn, error)

func secureAuthorityClient(client *http.Client, resolver authorityResolver) (*http.Client, error) {
	secured := *client
	var transport *http.Transport
	switch current := client.Transport.(type) {
	case nil:
		defaultTransport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return nil, errors.New("online specification checks require the standard default HTTP transport")
		}
		transport = defaultTransport.Clone()
	case *http.Transport:
		transport = current.Clone()
	default:
		return nil, errors.New("online specification checks require an HTTP transport with pinned destination dialing")
	}
	dial := authorityDialer(transport.DialContext)
	if dial == nil {
		dial = (&net.Dialer{}).DialContext
	}
	transport.Proxy = nil
	transport.DialTLSContext = nil
	transport.DialContext = dialResolvedAuthority(resolver, dial)
	secured.Transport = transport
	return &secured, nil
}

func dialResolvedAuthority(resolver authorityResolver, dial authorityDialer) authorityDialer {
	return func(ctx context.Context, network, target string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(target)
		if err != nil {
			return nil, fmt.Errorf("split authority dial target: %w", err)
		}
		destination := &url.URL{Scheme: "https", Host: net.JoinHostPort(host, port)}
		addresses, err := resolveAuthorityDestination(ctx, resolver, destination)
		if err != nil {
			return nil, err
		}
		var dialErr error
		for _, address := range addresses {
			connection, currentErr := dial(ctx, network, net.JoinHostPort(address.String(), port))
			if currentErr == nil {
				return connection, nil
			}
			dialErr = errors.Join(dialErr, currentErr)
		}
		return nil, fmt.Errorf("dial authority destination: %w", dialErr)
	}
}

// Check validates the repository-level decision register, conformance matrix,
// source provenance, and resolved executable evidence when any module declares
// specification ownership.
func Check(root string, catalog inventory.Inventory) (Report, error) {
	report, _, err := check(root, catalog)
	if err == nil && report.Unresolved > 0 {
		err = fmt.Errorf("specification decisions require review: %d unresolved", report.Unresolved)
	}
	return report, err
}

func check(root string, catalog inventory.Inventory) (Report, []monitoring, error) {
	report := Report{}
	policies := make([]monitoring, 0, len(catalog.Modules))
	decisionOwners := make(map[string]string)
	for _, module := range catalog.Modules {
		if len(module.Specifications) == 0 {
			provenance := strings.TrimSpace(string(module.Provenance))
			if provenance != "" && provenance != "null" && provenance != "[]" {
				return Report{}, nil, fmt.Errorf("module %q has specification provenance but does not declare specifications", module.Directory)
			}
			hasArtifacts, err := hasSpecificationArtifacts(root, module.Directory)
			if err != nil {
				return Report{}, nil, fmt.Errorf("module %q specification discovery: %w", module.Directory, err)
			}
			if hasArtifacts {
				return Report{}, nil, fmt.Errorf("module %q has specification artifacts but does not declare specifications", module.Directory)
			}
			continue
		}
		if err := validateProvenance(root, module, catalog.Modules); err != nil {
			return Report{}, nil, fmt.Errorf("module %q specification provenance: %w", module.Directory, err)
		}
		moduleReport, policy, identifiers, err := checkModule(root, catalog, module)
		if err != nil {
			return Report{}, nil, fmt.Errorf("module %q: %w", module.Directory, err)
		}
		for _, identifier := range identifiers {
			if owner, exists := decisionOwners[identifier]; exists {
				return Report{}, nil, fmt.Errorf("specification decision identifier %q is shared by modules %q and %q", identifier, owner, module.Directory)
			}
			decisionOwners[identifier] = module.Directory
		}
		report.Modules += moduleReport.Modules
		report.Decisions += moduleReport.Decisions
		report.Unresolved += moduleReport.Unresolved
		policies = append(policies, policy)
	}
	return report, policies, nil
}

func hasSpecificationArtifacts(root, moduleDirectory string) (bool, error) {
	path := filepath.Join(root, modulePath(moduleDirectory, "specification"))
	_, err := os.Lstat(path)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	default:
		return false, err
	}
}

func checkModule(root string, catalog inventory.Inventory, module inventory.Module) (Report, monitoring, []string, error) {
	registerPath := modulePath(module.Directory, "docs/specification-decisions.md")
	matrixPath := modulePath(module.Directory, "specification/README.md")
	monitoringPath := modulePath(module.Directory, "specification/monitoring.json")
	decisionsPath := modulePath(module.Directory, "specification/decisions.json")
	conformancePath := modulePath(module.Directory, "specification/conformance.json")
	historyPath := modulePath(module.Directory, "specification/decision-history.json")
	policy, err := loadMonitoring(root, monitoringPath, module.Specifications, time.Now().UTC())
	if err != nil {
		return Report{}, monitoring{}, nil, err
	}
	register, err := repositoryfile.Read(root, registerPath, maximumDocumentSize)
	if err != nil {
		return Report{}, monitoring{}, nil, fmt.Errorf("read specification decision register: %w", err)
	}
	matrix, err := repositoryfile.Read(root, matrixPath, maximumDocumentSize)
	if err != nil {
		return Report{}, monitoring{}, nil, fmt.Errorf("read specification conformance matrix: %w", err)
	}
	evidence, err := testEvidence(root, module.Directory, catalog.Modules)
	if err != nil {
		return Report{}, monitoring{}, nil, err
	}
	decisionData, err := repositoryfile.Read(root, decisionsPath, maximumDocumentSize)
	if err != nil {
		return Report{}, monitoring{}, nil, fmt.Errorf("read specification decisions.json: %w", err)
	}
	conformanceData, err := repositoryfile.Read(root, conformancePath, maximumDocumentSize)
	if err != nil {
		return Report{}, monitoring{}, nil, fmt.Errorf("read specification conformance.json: %w", err)
	}
	decisions, err := loadDecisions(decisionData)
	if err != nil {
		return Report{}, monitoring{}, nil, err
	}
	conformance, err := loadConformance(conformanceData)
	if err != nil {
		return Report{}, monitoring{}, nil, err
	}
	historyData, err := repositoryfile.Read(root, historyPath, maximumDocumentSize)
	if err != nil {
		return Report{}, monitoring{}, nil, fmt.Errorf("read specification decision-history.json: %w", err)
	}
	var history decisionHistory
	if err := decodeStrictJSON(historyData, &history); err != nil {
		return Report{}, monitoring{}, nil, fmt.Errorf("decode specification decision-history.json: %w", err)
	}
	if history.SchemaVersion != 1 {
		return Report{}, monitoring{}, nil, errors.New("specification decision-history.json schema_version must be 1")
	}
	matrixIDs, err := matrixDecisions(matrix)
	if err != nil {
		return Report{}, monitoring{}, nil, err
	}
	if !decisionHeading.Match(register) {
		return Report{}, monitoring{}, nil, errors.New("specification decision register contains no stable decision entries")
	}
	report, identifiers, err := validateDecisionContract(root, module, catalog.Modules, policy, decisions, conformance, register, matrixIDs, evidence)
	if err != nil {
		return Report{}, monitoring{}, nil, err
	}
	if err := validateDecisionLedger(history, decisions.Decisions); err != nil {
		return Report{}, monitoring{}, nil, err
	}
	return report, policy, identifiers, nil
}

func modulePath(directory, relative string) string {
	if directory == "." || directory == "" {
		return relative
	}
	return filepath.Join(directory, relative)
}

func matrixDecisions(matrix []byte) (map[string]struct{}, error) {
	identifiers := decisionID.FindAllString(string(matrix), -1)
	seen := make(map[string]struct{}, len(identifiers))
	for _, identifier := range identifiers {
		if _, exists := seen[identifier]; exists {
			return nil, fmt.Errorf("conformance matrix contains duplicate specification decision %q", identifier)
		}
		seen[identifier] = struct{}{}
	}
	if len(seen) == 0 {
		return nil, errors.New("specification conformance matrix contains no decision identifiers")
	}
	return seen, nil
}

type decisionManifest struct {
	SchemaVersion int           `json:"schema_version"`
	ChangeControl changeControl `json:"change_control"`
	Decisions     []decision    `json:"decisions"`
}

type changeControl struct {
	README              string `json:"readme"`
	Conformance         string `json:"conformance"`
	Compatibility       string `json:"compatibility"`
	Contribution        string `json:"contribution"`
	Changelog           string `json:"changelog"`
	PullRequestTemplate string `json:"pull_request_template"`
}

type decision struct {
	ID                        string   `json:"id"`
	Title                     string   `json:"title"`
	Status                    string   `json:"status"`
	Owner                     string   `json:"owner"`
	Classification            string   `json:"classification"`
	DecisionScope             string   `json:"decision_scope"`
	Specification             string   `json:"specification"`
	Version                   string   `json:"version"`
	SourceAuthority           string   `json:"source_authority"`
	Section                   string   `json:"section"`
	RequirementStrength       string   `json:"requirement_strength"`
	Issue                     string   `json:"issue"`
	Interpretations           []string `json:"interpretations"`
	PeerBehavior              string   `json:"peer_behavior"`
	SelectedBehavior          string   `json:"selected_behavior"`
	Rationale                 string   `json:"rationale"`
	SecurityConsequences      string   `json:"security_consequences"`
	ResourceConsequences      string   `json:"resource_consequences"`
	CompatibilityConsequences string   `json:"compatibility_consequences"`
	WireConsequences          string   `json:"wire_consequences"`
	ExecutableEvidence        []string `json:"executable_evidence"`
	FixtureEvidence           []string `json:"fixture_evidence"`
	FuzzEvidence              []string `json:"fuzz_evidence"`
	InteroperabilityEvidence  []string `json:"interoperability_evidence"`
	DifferentialEvidence      []string `json:"differential_evidence,omitempty"`
	PublicAPIs                []string `json:"public_apis"`
	Documentation             []string `json:"documentation"`
	UpstreamStatus            string   `json:"upstream_status"`
	ReconsiderWhen            string   `json:"reconsider_when"`
	Replacement               string   `json:"replacement,omitempty"`
}

type conformanceManifest struct {
	SchemaVersion int                   `json:"schema_version"`
	Decisions     []conformanceDecision `json:"decisions"`
}

type conformanceDecision struct {
	ID                             string   `json:"id"`
	AuthoritativeSources           []string `json:"authoritative_sources"`
	ExecutableEvidence             []string `json:"executable_evidence"`
	Fixtures                       []string `json:"fixtures"`
	Fuzz                           []string `json:"fuzz"`
	InteroperabilityEvidence       []string `json:"interoperability_evidence,omitempty"`
	InteroperabilityClassification string   `json:"interoperability_classification,omitempty"`
	DifferentialEvidence           []string `json:"differential_evidence"`
	DifferentialClassification     string   `json:"differential_classification"`
	PublicBehavior                 []string `json:"public_behavior"`
}

var allowedClassifications = map[string]struct{}{
	"ambiguity": {}, "contradiction": {}, "omission": {}, "erratum": {},
	"implementation-defined behavior": {}, "optional behavior": {}, "interoperability policy": {},
}

var allowedDecisionScopes = map[string]struct{}{
	"normative": {}, "recommended": {}, "defensive": {}, "extension-specific": {},
	"transport-specific": {}, "application-policy": {},
}

var allowedRequirementStrengths = map[string]struct{}{
	"MUST": {}, "MUST NOT": {}, "REQUIRED": {}, "SHALL": {}, "SHALL NOT": {},
	"SHOULD": {}, "SHOULD NOT": {}, "RECOMMENDED": {}, "NOT RECOMMENDED": {},
	"MAY": {}, "OPTIONAL": {}, "not specified": {}, "informative": {},
}

var allowedDifferentialClassifications = map[string]struct{}{
	"local defect": {}, "peer defect": {}, "fixture defect": {}, "harness defect": {},
	"specification ambiguity": {}, "deliberate policy difference": {}, "maintained peer agreement": {},
}

var allowedInteroperabilityClassifications = map[string]struct{}{
	"official fixture agreement": {}, "provider agreement": {},
}

func loadDecisions(data []byte) (decisionManifest, error) {
	var manifest decisionManifest
	if err := decodeStrictJSON(data, &manifest); err != nil {
		return decisionManifest{}, fmt.Errorf("decode specification decisions.json: %w", err)
	}
	if manifest.SchemaVersion != 1 {
		return decisionManifest{}, errors.New("specification decisions.json schema_version must be 1")
	}
	if len(manifest.Decisions) == 0 {
		return decisionManifest{}, errors.New("specification decisions.json contains no decisions")
	}
	return manifest, nil
}

func loadConformance(data []byte) (conformanceManifest, error) {
	var manifest conformanceManifest
	if err := decodeStrictJSON(data, &manifest); err != nil {
		return conformanceManifest{}, fmt.Errorf("decode specification conformance.json: %w", err)
	}
	if manifest.SchemaVersion != 1 {
		return conformanceManifest{}, errors.New("specification conformance.json schema_version must be 1")
	}
	if len(manifest.Decisions) == 0 {
		return conformanceManifest{}, errors.New("specification conformance.json contains no decisions")
	}
	return manifest, nil
}

func decodeStrictJSON(data []byte, target any) error {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return fmt.Errorf("decode trailing data: %w", err)
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	return walkJSONValue(json.NewDecoder(bytes.NewReader(data)))
}

func walkJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	if delimiter == '{' {
		keys := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key := fmt.Sprint(keyToken)
			if _, exists := keys[key]; exists {
				return fmt.Errorf("duplicate JSON object key %q", key)
			}
			keys[key] = struct{}{}
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
	} else {
		for decoder.More() {
			switch err := walkJSONValue(decoder); err {
			case nil:
			default:
				return err
			}
		}
	}
	_, err = decoder.Token()
	return err
}

func validateDecisionContract(
	root string,
	module inventory.Module,
	modules []inventory.Module,
	policy monitoring,
	manifest decisionManifest,
	conformance conformanceManifest,
	register []byte,
	matrixIDs map[string]struct{},
	evidence map[string]struct{},
) (Report, []string, error) {
	authorities := make(map[string]authority, len(policy.Authorities))
	for _, item := range policy.Authorities {
		authorities[item.ID] = item
	}
	registerEntries := make(map[string]string)
	registerBodies := make(map[string]string)
	registerMatches := decisionHeading.FindAllSubmatchIndex(register, -1)
	for _, match := range registerMatches {
		identifier := string(register[match[2]:match[3]])
		if _, exists := registerEntries[identifier]; exists {
			return Report{}, nil, fmt.Errorf("duplicate specification decision identifier %q", identifier)
		}
		registerEntries[identifier] = strings.TrimSpace(string(register[match[4]:match[5]]))
		end := len(register)
		if next := levelTwoHeading.FindIndex(register[match[1]:]); next != nil {
			end = match[1] + next[0]
		}
		registerBodies[identifier] = string(register[match[0]:end])
	}
	if err := validateChangeControl(root, module, modules, manifest.ChangeControl, modulePath(module.Directory, "docs/specification-decisions.md"), manifest.Decisions); err != nil {
		return Report{}, nil, err
	}
	conformanceEntries := make(map[string]conformanceDecision, len(conformance.Decisions))
	for _, item := range conformance.Decisions {
		if !decisionID.MatchString(item.ID) || decisionID.FindString(item.ID) != item.ID {
			return Report{}, nil, fmt.Errorf("conformance decision has invalid identifier %q", item.ID)
		}
		if _, exists := conformanceEntries[item.ID]; exists {
			return Report{}, nil, fmt.Errorf("duplicate conformance decision %q", item.ID)
		}
		conformanceEntries[item.ID] = item
	}
	report := Report{Modules: 1}
	identifiers := make([]string, 0, len(manifest.Decisions))
	decisions := make(map[string]decision, len(manifest.Decisions))
	for _, item := range manifest.Decisions {
		if err := validateDecision(root, module, modules, item, authorities, evidence); err != nil {
			return Report{}, nil, err
		}
		if _, exists := decisions[item.ID]; exists {
			return Report{}, nil, fmt.Errorf("duplicate specification decision %q", item.ID)
		}
		decisions[item.ID] = item
		identifiers = append(identifiers, item.ID)
		if title, exists := registerEntries[item.ID]; !exists {
			return Report{}, nil, fmt.Errorf("specification decision %q is missing from the Markdown register", item.ID)
		} else if title != item.Title {
			return Report{}, nil, fmt.Errorf("specification decision %q title differs between JSON and Markdown", item.ID)
		}
		if err := validateMarkdownDecision(item, authorities[item.SourceAuthority], registerBodies[item.ID]); err != nil {
			return Report{}, nil, err
		}
		if _, exists := matrixIDs[item.ID]; !exists {
			return Report{}, nil, fmt.Errorf("specification decision %q is missing from the conformance matrix", item.ID)
		}
		row, exists := conformanceEntries[item.ID]
		if !exists {
			return Report{}, nil, fmt.Errorf("specification decision %q is missing from conformance.json", item.ID)
		}
		if err := validateAdditionalAuthoritativeSourceDocumentation(item, row, authorities, registerBodies[item.ID]); err != nil {
			return Report{}, nil, err
		}
		if err := validateConformance(item, row, authorities, evidence); err != nil {
			return Report{}, nil, err
		}
		if item.Status == "unresolved" {
			report.Unresolved++
		}
		report.Decisions++
	}
	for identifier, item := range decisions {
		if item.Status == "superseded" {
			if item.Replacement == "" || item.Replacement == identifier {
				return Report{}, nil, fmt.Errorf("superseded specification decision %q has no replacement", identifier)
			}
			if _, exists := decisions[item.Replacement]; !exists {
				return Report{}, nil, fmt.Errorf("superseded specification decision %q references unknown replacement %q", identifier, item.Replacement)
			}
			if !documentLinksToDecision([]byte(registerBodies[identifier]), item.Replacement, decisions[item.Replacement].Title) {
				return Report{}, nil, fmt.Errorf("superseded specification decision %q has no Markdown replacement link to %q", identifier, item.Replacement)
			}
		}
	}
	for identifier := range registerEntries {
		if _, exists := decisions[identifier]; !exists {
			return Report{}, nil, fmt.Errorf("markdown register contains unknown specification decision %q", identifier)
		}
	}
	for identifier := range matrixIDs {
		if _, exists := decisions[identifier]; !exists {
			return Report{}, nil, fmt.Errorf("conformance matrix references unknown specification decision %q", identifier)
		}
	}
	for identifier := range conformanceEntries {
		if _, exists := decisions[identifier]; !exists {
			return Report{}, nil, fmt.Errorf("conformance.json references unknown specification decision %q", identifier)
		}
	}
	if err := validateDecisionHistory(root, manifest.ChangeControl.Changelog, manifest.Decisions); err != nil {
		return Report{}, nil, err
	}
	return report, identifiers, nil
}

type decisionHistory struct {
	SchemaVersion int                    `json:"schema_version"`
	Entries       []decisionHistoryEntry `json:"entries"`
}
type decisionHistoryEntry struct {
	ID      string   `json:"id"`
	Digests []string `json:"digests"`
}

func validateDecisionLedger(history decisionHistory, current []decision) error {
	entries := make(map[string]decisionHistoryEntry, len(history.Entries))
	for _, entry := range history.Entries {
		if entry.ID == "" || len(entry.Digests) == 0 {
			return errors.New("specification decision history has an incomplete entry")
		}
		if _, exists := entries[entry.ID]; exists {
			return fmt.Errorf("specification decision history repeats %q", entry.ID)
		}
		for _, digest := range entry.Digests {
			if !sha256Value.MatchString(digest) {
				return fmt.Errorf("specification decision history %q has an invalid digest", entry.ID)
			}
		}
		entries[entry.ID] = entry
	}
	currentIDs := make(map[string]struct{}, len(current))
	for _, item := range current {
		currentIDs[item.ID] = struct{}{}
		entry, exists := entries[item.ID]
		if !exists || !slicesContains(entry.Digests, decisionDigest(item)) {
			return fmt.Errorf("specification decision history does not contain current decision %q", item.ID)
		}
	}
	for identifier := range entries {
		if _, exists := currentIDs[identifier]; !exists {
			return fmt.Errorf("specification decision history preserves %q but the current register does not", identifier)
		}
	}
	return nil
}

func validateDecisionHistory(root, path string, decisions []decision) error {
	data, err := repositoryfile.Read(root, path, maximumDocumentSize)
	if err != nil {
		return err
	}
	current := make(map[string]struct{}, len(decisions))
	for _, item := range decisions {
		current[item.ID] = struct{}{}
		if !bytes.Contains(data, []byte(item.ID+" sha256:"+decisionDigest(item))) {
			return fmt.Errorf("specification decision %q has no changelog record for its current content", item.ID)
		}
	}
	for _, identifier := range decisionID.FindAllString(string(data), -1) {
		if _, exists := current[identifier]; !exists {
			return fmt.Errorf("specification changelog preserves decision %q but the current register does not", identifier)
		}
	}
	return nil
}

func validateMarkdownDecision(item decision, source authority, body string) error {
	values := []string{item.Status, item.Owner, item.Classification, item.DecisionScope, item.Specification, item.Version,
		item.SourceAuthority, source.URL, item.Section, item.RequirementStrength, item.Issue, item.PeerBehavior,
		item.SelectedBehavior, item.Rationale, item.SecurityConsequences, item.ResourceConsequences,
		item.CompatibilityConsequences, item.WireConsequences, item.UpstreamStatus, item.ReconsiderWhen}
	for _, list := range [][]string{item.Interpretations, item.ExecutableEvidence, item.FixtureEvidence, item.FuzzEvidence,
		item.InteroperabilityEvidence, item.DifferentialEvidence, item.PublicAPIs, item.Documentation} {
		values = append(values, list...)
	}
	for _, value := range values {
		if !strings.Contains(body, value) {
			return fmt.Errorf("specification decision %q Markdown entry omits exact value %q", item.ID, value)
		}
	}
	return nil
}

func validateChangeControl(root string, module inventory.Module, modules []inventory.Module, control changeControl, registerPath string, _ []decision) error {
	paths := map[string]string{
		"readme": control.README, "conformance": control.Conformance, "compatibility": control.Compatibility,
		"contribution": control.Contribution, "changelog": control.Changelog, "pull_request_template": control.PullRequestTemplate,
	}
	for name, path := range paths {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("specification change control has empty %s path", name)
		}
		if !repositoryPath(path) {
			return fmt.Errorf("specification change control %s path %q is invalid", name, path)
		}
	}
	for name, path := range map[string]string{
		"readme": control.README, "conformance": control.Conformance,
		"compatibility": control.Compatibility, "changelog": control.Changelog,
	} {
		if !moduleOwnsPath(module.Directory, path, modules) {
			return fmt.Errorf("specification change control %s %q is outside its module", name, path)
		}
	}
	for name, path := range map[string]string{
		"contribution": control.Contribution, "pull_request_template": control.PullRequestTemplate,
	} {
		if !moduleOwnsOrSharesPath(module.Directory, path, modules) {
			return fmt.Errorf("specification change control %s %q is outside its module and repository-shared paths", name, path)
		}
	}
	for name, path := range map[string]string{
		"readme": control.README, "conformance": control.Conformance, "compatibility": control.Compatibility,
		"contribution": control.Contribution, "changelog": control.Changelog,
	} {
		data, err := repositoryfile.Read(root, path, maximumDocumentSize)
		if err != nil {
			return fmt.Errorf("read specification change control %s %q: %w", name, path, err)
		}
		if !documentLinksTo(path, data, registerPath) {
			return fmt.Errorf("specification change control %s %q does not link the decision register", name, path)
		}
	}
	template, err := repositoryfile.Read(root, control.PullRequestTemplate, maximumDocumentSize)
	if err != nil {
		return fmt.Errorf("read specification pull request template %q: %w", control.PullRequestTemplate, err)
	}
	lower := strings.ToLower(string(template))
	if !strings.Contains(lower, "## specification decisions") || !strings.Contains(lower, "decision identifier") ||
		!strings.Contains(lower, "compatibility") || !strings.Contains(lower, "changelog") || !strings.Contains(lower, "superseded") {
		return fmt.Errorf("specification pull request template %q lacks the required decision review section", control.PullRequestTemplate)
	}
	return nil
}

func decisionDigest(item decision) string {
	data, _ := json.Marshal(item)
	var canonical any
	_ = json.Unmarshal(data, &canonical)
	data, _ = json.Marshal(canonical)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func repositoryPath(path string) bool {
	clean := filepath.Clean(path)
	return !filepath.IsAbs(clean) && clean != "." && clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator))
}

func documentLinksTo(documentPath string, data []byte, targetPath string) bool {
	for _, match := range markdownLink.FindAllSubmatch(data, -1) {
		parsed, err := url.Parse(markdownLinkDestination(match))
		if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.Path == "" {
			continue
		}
		resolved := filepath.Clean(filepath.Join(filepath.Dir(documentPath), filepath.FromSlash(parsed.Path)))
		if resolved == filepath.Clean(targetPath) {
			return true
		}
	}
	return false
}

func documentLinksToDecision(data []byte, identifier, title string) bool {
	expected := decisionHeadingAnchor(identifier, title)
	for _, match := range markdownLink.FindAllSubmatch(data, -1) {
		parsed, err := url.Parse(markdownLinkDestination(match))
		if err == nil && parsed.Scheme == "" && parsed.Host == "" && parsed.Path == "" && strings.EqualFold(parsed.Fragment, expected) {
			return true
		}
	}
	return false
}

func decisionHeadingAnchor(identifier, title string) string {
	var anchor strings.Builder
	for _, value := range strings.ToLower(identifier + " " + title) {
		if unicode.IsLetter(value) || unicode.IsDigit(value) || value == '-' || value == '_' {
			anchor.WriteRune(value)
		} else if unicode.IsSpace(value) {
			anchor.WriteByte('-')
		}
	}
	return anchor.String()
}

func markdownLinkDestination(match [][]byte) string {
	if len(match[1]) > 0 {
		return string(match[1])
	}
	return string(match[2])
}

func validateDecision(root string, module inventory.Module, modules []inventory.Module, item decision, authorities map[string]authority, evidence map[string]struct{}) error {
	if !decisionID.MatchString(item.ID) || decisionID.FindString(item.ID) != item.ID {
		return fmt.Errorf("specification decision has invalid identifier %q", item.ID)
	}
	for name, value := range map[string]string{
		"title": item.Title, "owner": item.Owner, "specification": item.Specification, "version": item.Version,
		"source_authority": item.SourceAuthority, "section": item.Section, "requirement_strength": item.RequirementStrength, "issue": item.Issue,
		"peer_behavior": item.PeerBehavior, "selected_behavior": item.SelectedBehavior,
		"rationale": item.Rationale, "security_consequences": item.SecurityConsequences,
		"resource_consequences": item.ResourceConsequences, "compatibility_consequences": item.CompatibilityConsequences,
		"wire_consequences": item.WireConsequences, "upstream_status": item.UpstreamStatus,
		"reconsider_when": item.ReconsiderWhen,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("specification decision %q has empty %s", item.ID, name)
		}
	}
	if item.Status != "resolved" && item.Status != "unresolved" && item.Status != "superseded" {
		return fmt.Errorf("specification decision %q has invalid status %q", item.ID, item.Status)
	}
	if _, exists := allowedClassifications[item.Classification]; !exists {
		return fmt.Errorf("specification decision %q has invalid classification %q", item.ID, item.Classification)
	}
	if _, exists := allowedDecisionScopes[item.DecisionScope]; !exists {
		return fmt.Errorf("specification decision %q has invalid decision_scope %q", item.ID, item.DecisionScope)
	}
	if _, exists := allowedRequirementStrengths[item.RequirementStrength]; !exists {
		return fmt.Errorf("specification decision %q has invalid requirement_strength %q", item.ID, item.RequirementStrength)
	}
	authorityItem, exists := authorities[item.SourceAuthority]
	if !exists || (authorityItem.Kind != "specification" && authorityItem.Kind != "registry" && authorityItem.Kind != "extension" && authorityItem.Kind != "recommendation") {
		return fmt.Errorf("specification decision %q references unknown source authority %q", item.ID, item.SourceAuthority)
	}
	mandatoryStrength := item.RequirementStrength == "REQUIRED" || strings.Contains(item.RequirementStrength, "MUST") || strings.Contains(item.RequirementStrength, "SHALL")
	if mandatoryStrength && item.DecisionScope != "normative" ||
		authorityItem.Kind == "recommendation" && item.DecisionScope == "normative" {
		return fmt.Errorf("specification decision %q has inconsistent authority, scope, and requirement_strength", item.ID)
	}
	if !slicesContains(authorityItem.Specifications, item.Specification) {
		return fmt.Errorf("specification decision %q source authority does not cover %q", item.ID, item.Specification)
	}
	if item.Version != authorityItem.Version {
		return fmt.Errorf("specification decision %q version %q does not match authority version %q", item.ID, item.Version, authorityItem.Version)
	}
	if err := validateTextList(item.ID, "interpretations", item.Interpretations); err != nil {
		return err
	}
	for name, values := range map[string][]string{
		"public_apis": item.PublicAPIs, "documentation": item.Documentation,
	} {
		if err := validateTextList(item.ID, name, values); err != nil {
			return err
		}
	}
	for name, values := range map[string][]string{
		"executable_evidence": item.ExecutableEvidence, "fixture_evidence": item.FixtureEvidence,
		"fuzz_evidence": item.FuzzEvidence, "interoperability_evidence": item.InteroperabilityEvidence,
		"differential_evidence": item.DifferentialEvidence,
	} {
		if err := validateOptionalTextList(item.ID, name, values); err != nil {
			return err
		}
	}
	if item.Status == "resolved" && len(item.ExecutableEvidence) == 0 {
		return fmt.Errorf("resolved specification decision %q has no executable evidence", item.ID)
	}
	if item.Status == "resolved" && !hasEvidencePrefix(item.ExecutableEvidence, "Test") {
		return fmt.Errorf("resolved specification decision %q has no Go test evidence", item.ID)
	}
	for _, name := range item.ExecutableEvidence {
		if _, exists := evidence[name]; !exists && item.Status != "superseded" {
			return fmt.Errorf("specification decision %q references missing executable evidence %q", item.ID, name)
		}
	}
	for _, name := range item.FuzzEvidence {
		if !strings.HasPrefix(name, "Fuzz") {
			return fmt.Errorf("specification decision %q fuzz evidence %q is not a Go fuzz target", item.ID, name)
		}
		if _, exists := evidence[name]; !exists && item.Status != "superseded" {
			return fmt.Errorf("specification decision %q references missing fuzz evidence %q", item.ID, name)
		}
	}
	for name, paths := range map[string][]string{
		"fixture evidence": item.FixtureEvidence, "interoperability evidence": item.InteroperabilityEvidence,
		"differential evidence": item.DifferentialEvidence,
	} {
		var err error
		if item.Status == "superseded" {
			err = validateOwnedPaths(module, modules, item.ID, name, paths)
		} else {
			err = validateOwnedFiles(root, module, modules, item.ID, name, paths)
		}
		if err != nil {
			return err
		}
	}
	if err := validateOwnedFiles(root, module, modules, item.ID, "documentation", item.Documentation); err != nil {
		return err
	}
	return nil
}

func hasEvidencePrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func validateOwnedFiles(root string, module inventory.Module, modules []inventory.Module, identifier, name string, paths []string) error {
	if err := validateOwnedPaths(module, modules, identifier, name, paths); err != nil {
		return err
	}
	for _, path := range paths {
		if _, err := repositoryfile.Read(root, path, maximumDocumentSize); err != nil {
			return fmt.Errorf("specification decision %q %s %q: %w", identifier, name, path, err)
		}
	}
	return nil
}

func validateOwnedPaths(module inventory.Module, modules []inventory.Module, identifier, name string, paths []string) error {
	for _, path := range paths {
		if !moduleOwnsPath(module.Directory, path, modules) {
			return fmt.Errorf("specification decision %q %s %q is outside its module", identifier, name, path)
		}
	}
	return nil
}

func validateConformance(item decision, row conformanceDecision, authorities map[string]authority, evidence map[string]struct{}) error {
	for name, values := range map[string][]string{
		"authoritative_sources": row.AuthoritativeSources, "public_behavior": row.PublicBehavior,
	} {
		if err := validateTextList(item.ID, "conformance "+name, values); err != nil {
			return err
		}
	}
	for name, values := range map[string][]string{
		"executable_evidence": row.ExecutableEvidence, "fixtures": row.Fixtures,
		"fuzz": row.Fuzz, "interoperability_evidence": row.InteroperabilityEvidence,
		"differential_evidence": row.DifferentialEvidence,
	} {
		if err := validateOptionalTextList(item.ID, "conformance "+name, values); err != nil {
			return err
		}
	}
	if !slicesContains(row.AuthoritativeSources, item.SourceAuthority) {
		return fmt.Errorf("conformance decision %q omits source authority %q", item.ID, item.SourceAuthority)
	}
	for _, identifier := range row.AuthoritativeSources {
		source, exists := authorities[identifier]
		if !exists || (source.Kind != "specification" && source.Kind != "registry" && source.Kind != "extension" && source.Kind != "recommendation") {
			return fmt.Errorf("conformance decision %q references unknown authoritative source %q", item.ID, identifier)
		}
		if identifier == item.SourceAuthority && !slicesContains(source.Specifications, item.Specification) {
			return fmt.Errorf("conformance decision %q primary authoritative source %q does not cover %q", item.ID, identifier, item.Specification)
		}
	}
	expectedInteroperability := item.InteroperabilityEvidence
	expectedDifferential := item.DifferentialEvidence
	if len(row.InteroperabilityEvidence) == 0 && len(expectedDifferential) == 0 {
		// Preserve compatibility with schema v1 records that used the generic
		// decision field for maintained-peer differential evidence.
		expectedInteroperability = nil
		expectedDifferential = item.InteroperabilityEvidence
	}
	for name, pair := range map[string][2][]string{
		"executable evidence":       {item.ExecutableEvidence, row.ExecutableEvidence},
		"fixtures":                  {item.FixtureEvidence, row.Fixtures},
		"fuzz evidence":             {item.FuzzEvidence, row.Fuzz},
		"interoperability evidence": {expectedInteroperability, row.InteroperabilityEvidence},
		"differential evidence":     {expectedDifferential, row.DifferentialEvidence},
	} {
		if !sameStringSet(pair[0], pair[1]) {
			return fmt.Errorf("conformance decision %q %s differ from decisions.json", item.ID, name)
		}
	}
	switch {
	case len(row.InteroperabilityEvidence) == 0:
		if row.InteroperabilityClassification != "" && row.InteroperabilityClassification != "not assessed" {
			return fmt.Errorf("conformance decision %q without interoperability evidence must be not assessed", item.ID)
		}
	case row.InteroperabilityClassification == "" || row.InteroperabilityClassification == "not assessed":
		return fmt.Errorf("conformance decision %q with interoperability evidence must have an assessed classification", item.ID)
	default:
		if _, exists := allowedInteroperabilityClassifications[row.InteroperabilityClassification]; !exists {
			return fmt.Errorf("conformance decision %q has invalid interoperability classification %q", item.ID, row.InteroperabilityClassification)
		}
	}
	switch {
	case len(row.DifferentialEvidence) == 0:
		if row.DifferentialClassification != "not assessed" {
			return fmt.Errorf("conformance decision %q without differential evidence must be not assessed", item.ID)
		}
	case row.DifferentialClassification == "not assessed":
		return fmt.Errorf("conformance decision %q with differential evidence must have an assessed classification", item.ID)
	default:
		if _, exists := allowedDifferentialClassifications[row.DifferentialClassification]; !exists {
			return fmt.Errorf("conformance decision %q has invalid differential classification %q", item.ID, row.DifferentialClassification)
		}
	}
	for _, name := range row.ExecutableEvidence {
		if _, exists := evidence[name]; !exists && item.Status != "superseded" {
			return fmt.Errorf("conformance decision %q references missing executable evidence %q", item.ID, name)
		}
	}
	return nil
}

func validateAdditionalAuthoritativeSourceDocumentation(item decision, row conformanceDecision, authorities map[string]authority, body string) error {
	for _, identifier := range row.AuthoritativeSources {
		if identifier == item.SourceAuthority {
			continue
		}
		source, exists := authorities[identifier]
		if !exists {
			return fmt.Errorf("specification decision %q references unknown authoritative source %q in an additional binding", item.ID, identifier)
		}
		record := documentedAuthority{ID: source.ID, Version: source.Version, URL: source.URL, Specifications: source.Specifications}
		encoded, _ := json.Marshal(record)
		expected := "`" + string(encoded) + "`"
		if !strings.Contains(body, expected) {
			return fmt.Errorf("specification decision %q Markdown entry omits additional authoritative source %q record %s", item.ID, identifier, expected)
		}
	}
	return nil
}

type documentedAuthority struct {
	ID             string   `json:"id"`
	Version        string   `json:"version"`
	URL            string   `json:"url"`
	Specifications []string `json:"specifications"`
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[string]struct{}, len(left))
	for _, value := range left {
		seen[value] = struct{}{}
	}
	for _, value := range right {
		if _, exists := seen[value]; !exists {
			return false
		}
	}
	return true
}

func validateTextList(identifier, name string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("specification decision %q has empty %s", identifier, name)
	}
	return validateOptionalTextList(identifier, name, values)
}

func validateOptionalTextList(identifier, name string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("specification decision %q has blank %s entry", identifier, name)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("specification decision %q repeats %s entry %q", identifier, name, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func slicesContains(values []string, expected string) bool {
	return slices.Contains(values, expected)
}

type monitoring struct {
	SchemaVersion      int         `json:"schema_version"`
	ReviewedAt         string      `json:"reviewed_at"`
	ReviewIntervalDays int         `json:"review_interval_days"`
	Authorities        []authority `json:"authorities"`
}

type authority struct {
	ID                string   `json:"id"`
	Kind              string   `json:"kind"`
	Version           string   `json:"version"`
	URL               string   `json:"url"`
	SHA256            string   `json:"sha256,omitempty"`
	Access            string   `json:"access,omitempty"`
	ExpectedStatus    int      `json:"expected_status,omitempty"`
	UnavailableReason string   `json:"unavailable_reason,omitempty"`
	Specifications    []string `json:"specifications"`
}

func loadMonitoring(root, relative string, specifications []string, now time.Time) (monitoring, error) {
	data, err := repositoryfile.Read(root, relative, maximumDocumentSize)
	if err != nil {
		return monitoring{}, fmt.Errorf("read specification authority monitoring: %w", err)
	}
	var policy monitoring
	if err := decodeStrictJSON(data, &policy); err != nil {
		return monitoring{}, fmt.Errorf("decode specification authority monitoring: %w", err)
	}
	if policy.SchemaVersion != 1 {
		return monitoring{}, errors.New("specification authority monitoring schema_version must be 1")
	}
	reviewed, err := time.Parse("2006-01-02", policy.ReviewedAt)
	if err != nil {
		return monitoring{}, errors.New("specification authority monitoring reviewed_at must be an ISO date")
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if reviewed.After(today) {
		return monitoring{}, errors.New("specification authority monitoring reviewed_at is in the future")
	}
	if policy.ReviewIntervalDays < 1 || policy.ReviewIntervalDays > 90 {
		return monitoring{}, errors.New("specification authority review interval must be between 1 and 90 days")
	}
	if reviewed.AddDate(0, 0, policy.ReviewIntervalDays).Before(today) {
		return monitoring{}, errors.New("specification authority review is stale")
	}
	if len(policy.Authorities) == 0 {
		return monitoring{}, errors.New("specification authority monitoring contains no authorities")
	}
	if len(policy.Authorities) > maximumAuthorities {
		return monitoring{}, fmt.Errorf("specification authority monitoring contains too many authorities: %d exceeds %d", len(policy.Authorities), maximumAuthorities)
	}
	seen := make(map[string]struct{}, len(policy.Authorities))
	declared := make(map[string]struct{}, len(specifications))
	coverage := make(map[string]struct{ source, change bool }, len(specifications))
	for _, specification := range specifications {
		if specification == "" {
			return monitoring{}, errors.New("specification authority monitoring received an empty specification declaration")
		}
		if _, exists := declared[specification]; exists {
			return monitoring{}, fmt.Errorf("duplicate declared specification %q", specification)
		}
		declared[specification] = struct{}{}
		coverage[specification] = struct{ source, change bool }{}
	}
	for index, item := range policy.Authorities {
		if _, exists := seen[item.ID]; item.ID == "" || exists {
			return monitoring{}, fmt.Errorf("specification authority %d has an empty or duplicate identifier", index)
		}
		seen[item.ID] = struct{}{}
		parsed, parseErr := url.Parse(item.URL)
		if parseErr != nil || !validPinnedHTTPSURL(parsed) {
			return monitoring{}, fmt.Errorf("specification authority %q is not an exact HTTPS authority", item.ID)
		}
		isSource := false
		isChange := false
		switch item.Kind {
		case "errata", "releases":
			isChange = true
		case "specification", "registry", "extension", "recommendation":
			isSource = true
		default:
			return monitoring{}, fmt.Errorf("specification authority %q has unsupported kind %q", item.ID, item.Kind)
		}
		switch item.Access {
		case "", "public":
			if !sha256Value.MatchString(item.SHA256) || item.ExpectedStatus != 0 || strings.TrimSpace(item.UnavailableReason) != "" {
				return monitoring{}, fmt.Errorf("public specification authority %q is not an exact content pin", item.ID)
			}
		case "restricted":
			if !isSource || item.SHA256 != "" || !restrictedAuthorityStatus(item.ExpectedStatus) || strings.TrimSpace(item.UnavailableReason) == "" {
				return monitoring{}, fmt.Errorf("restricted specification authority %q must be a source with no content digest, a pinned denial status, and an unavailable reason", item.ID)
			}
		default:
			return monitoring{}, fmt.Errorf("specification authority %q has unsupported access %q", item.ID, item.Access)
		}
		if isSource && strings.TrimSpace(item.Version) == "" {
			return monitoring{}, fmt.Errorf("specification source authority %q has no exact version", item.ID)
		}
		if len(item.Specifications) == 0 {
			return monitoring{}, fmt.Errorf("specification authority %q has no specification bindings", item.ID)
		}
		itemSpecifications := make(map[string]struct{}, len(item.Specifications))
		for _, specification := range item.Specifications {
			if _, duplicate := itemSpecifications[specification]; duplicate {
				return monitoring{}, fmt.Errorf("specification authority %q repeats specification %q", item.ID, specification)
			}
			itemSpecifications[specification] = struct{}{}
			if _, exists := declared[specification]; !exists {
				return monitoring{}, fmt.Errorf("specification authority %q references undeclared specification %q", item.ID, specification)
			}
			state := coverage[specification]
			state.source = state.source || isSource
			state.change = state.change || isChange
			coverage[specification] = state
		}
	}
	for specification, state := range coverage {
		if !state.source || !state.change {
			return monitoring{}, fmt.Errorf("declared specification %q lacks source and change authorities", specification)
		}
	}
	return policy, nil
}

func restrictedAuthorityStatus(status int) bool {
	return status == http.StatusUnauthorized || status == http.StatusForbidden || status == http.StatusUnavailableForLegalReasons
}

func validateProvenance(root string, module inventory.Module, modules []inventory.Module) error {
	var paths []string
	if len(module.Provenance) == 0 || string(module.Provenance) == "null" {
		return errors.New("source manifest is not declared")
	}
	if err := json.Unmarshal(module.Provenance, &paths); err != nil {
		return fmt.Errorf("decode source manifests: %w", err)
	}
	if len(paths) == 0 {
		return errors.New("source manifest is not declared")
	}
	for _, path := range paths {
		if !moduleOwnsPath(module.Directory, path, modules) {
			return fmt.Errorf("source manifest %q is outside module %q", path, module.Directory)
		}
		data, err := repositoryfile.Read(root, path, maximumDocumentSize)
		if err != nil {
			return fmt.Errorf("read %q: %w", path, err)
		}
		if err := validateSourceManifest(path, data); err != nil {
			return fmt.Errorf("validate %q: %w", path, err)
		}
	}
	return nil
}

func moduleOwnsPath(moduleDirectory, path string, modules []inventory.Module) bool {
	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return false
	}
	if moduleDirectory != "." && moduleDirectory != "" {
		moduleRoot := filepath.Clean(moduleDirectory)
		return clean != moduleRoot && strings.HasPrefix(clean, moduleRoot+string(filepath.Separator))
	}
	for _, module := range modules {
		if module.Directory == "." || module.Directory == "" {
			continue
		}
		nested := filepath.Clean(module.Directory)
		if clean == nested || strings.HasPrefix(clean, nested+string(filepath.Separator)) {
			return false
		}
	}
	return true
}

func moduleOwnsOrSharesPath(moduleDirectory, path string, modules []inventory.Module) bool {
	if moduleOwnsPath(moduleDirectory, path, modules) {
		return true
	}
	if !repositoryPath(path) {
		return false
	}
	clean := filepath.Clean(path)
	for _, module := range modules {
		if module.Directory == "." || module.Directory == "" {
			continue
		}
		nested := filepath.Clean(module.Directory)
		if clean == nested || strings.HasPrefix(clean, nested+string(filepath.Separator)) {
			return false
		}
	}
	return true
}

func validateSourceManifest(path string, data []byte) error {
	switch filepath.Ext(path) {
	case ".tsv":
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) < 2 {
			return errors.New("source manifest contains no pinned sources")
		}
		headings := strings.Split(lines[0], "\t")
		columns := make(map[string]int, len(headings))
		for index, heading := range headings {
			if _, exists := columns[heading]; exists {
				return fmt.Errorf("source manifest has duplicate %s column", heading)
			}
			columns[heading] = index
		}
		for _, required := range []string{"id", "version", "url", "sha256", "status"} {
			if _, exists := columns[required]; !exists {
				return fmt.Errorf("source manifest is missing %s column", required)
			}
		}
		for lineNumber, line := range lines[1:] {
			fields := strings.Split(line, "\t")
			if len(fields) != len(headings) {
				return fmt.Errorf("source row %d has %d fields, want %d", lineNumber+2, len(fields), len(headings))
			}
			parsed, err := url.Parse(fields[columns["url"]])
			if fields[columns["id"]] == "" || fields[columns["version"]] == "" || err != nil || !validPinnedHTTPSURL(parsed) ||
				!sha256Value.MatchString(fields[columns["sha256"]]) || fields[columns["status"]] != "pinned" {
				return fmt.Errorf("source row %d is not an exact HTTPS pin", lineNumber+2)
			}
		}
		return nil
	case ".json":
		if err := rejectDuplicateJSONKeys(data); err != nil {
			return fmt.Errorf("decode JSON source manifest: %w", err)
		}
		var value any
		_ = json.Unmarshal(data, &value) // The duplicate-key walk already parsed the complete value.
		pins, err := validateJSONPins(value)
		if err != nil {
			return err
		}
		if pins == 0 {
			return errors.New("JSON source manifest contains no integrity pins")
		}
		return nil
	default:
		return fmt.Errorf("unsupported source manifest format %q", filepath.Ext(path))
	}
}

func validPinnedHTTPSURL(parsed *url.URL) bool {
	return parsed != nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil && parsed.Fragment == ""
}

func validateJSONPins(value any) (int, error) {
	switch typed := value.(type) {
	case map[string]any:
		pins := 0
		for key, item := range typed {
			if strings.HasSuffix(strings.ToLower(key), "sha256") {
				digest, ok := item.(string)
				if !ok || !sha256Value.MatchString(digest) {
					return 0, fmt.Errorf("JSON source manifest has invalid %s integrity pin", key)
				}
				if !jsonPinHasIdentity(typed, key) {
					return 0, fmt.Errorf("JSON source manifest %s integrity pin has no source identity", key)
				}
				pins++
			}
			nested, err := validateJSONPins(item)
			if err != nil {
				return 0, err
			}
			pins += nested
		}
		return pins, nil
	case []any:
		pins := 0
		for _, item := range typed {
			nested, err := validateJSONPins(item)
			if err != nil {
				return 0, err
			}
			pins += nested
		}
		return pins, nil
	default:
		return 0, nil
	}
}

func jsonPinHasIdentity(object map[string]any, digestKey string) bool {
	prefix := strings.TrimSuffix(strings.ToLower(digestKey), "sha256")
	prefix = strings.TrimSuffix(prefix, "_")
	for key, value := range object {
		text, ok := value.(string)
		if ok && strings.TrimSpace(text) != "" {
			lower := strings.ToLower(key)
			for _, suffix := range []string{"id", "name", "version", "url", "path", "source", "commit", "tag"} {
				if lower == suffix || lower == prefix+"_"+suffix {
					return true
				}
			}
		}
	}
	return false
}

func testEvidence(root, moduleDirectory string, modules []inventory.Module) (map[string]struct{}, error) {
	moduleRoot := filepath.Join(root, filepath.Clean(moduleDirectory))
	excluded := make(map[string]struct{})
	for _, module := range modules {
		if module.Directory == "." || module.Directory == "" || module.Directory == moduleDirectory {
			continue
		}
		candidate := filepath.Join(root, filepath.Clean(module.Directory))
		if strings.HasPrefix(candidate+string(filepath.Separator), moduleRoot+string(filepath.Separator)) {
			excluded[candidate] = struct{}{}
		}
	}
	evidence := make(map[string]struct{})
	buildContext := build.Default
	for _, module := range modules {
		if module.Directory == moduleDirectory {
			buildContext.BuildTags = append([]string(nil), module.TestTags...)
		}
	}
	totalSize := 0
	err := filepath.WalkDir(moduleRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("inspect executable evidence: symlink is not allowed: %s", path)
		}
		if entry.IsDir() {
			if _, skip := excluded[path]; skip {
				return filepath.SkipDir
			}
			if path != moduleRoot && (entry.Name() == "vendor" || entry.Name() == "testdata" || strings.HasPrefix(entry.Name(), ".") || strings.HasPrefix(entry.Name(), "_")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		matched, err := buildContext.MatchFile(filepath.Dir(path), entry.Name())
		if err != nil {
			return fmt.Errorf("match executable specification evidence %q: %w", path, err)
		}
		if !matched {
			return nil
		}
		relative := strings.TrimPrefix(path, filepath.Clean(root)+string(filepath.Separator))
		data, err := repositoryfile.Read(root, relative, maximumEvidenceSize)
		if err != nil {
			return err
		}
		totalSize += len(data)
		if totalSize > maximumEvidenceSize {
			return errors.New("executable specification evidence exceeds maximum size")
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, data, 0)
		if err != nil {
			return fmt.Errorf("parse executable specification evidence %q: %w", relative, err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || function.Name == nil {
				continue
			}
			if isExecutableEvidence(file, function) {
				evidence[function.Name.Name] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect executable specification evidence: %w", err)
	}
	return evidence, nil
}

func isExecutableEvidence(file *ast.File, function *ast.FuncDecl) bool {
	if function.Recv != nil || function.Name == nil || function.Type.TypeParams != nil || function.Type.Results != nil || function.Type.Params == nil || len(function.Type.Params.List) != 1 || len(function.Type.Params.List[0].Names) > 1 {
		return false
	}
	var expected string
	switch {
	case executableName(function.Name.Name, "Test"):
		expected = "T"
	case executableName(function.Name.Name, "Fuzz"):
		expected = "F"
	case executableName(function.Name.Name, "Benchmark"):
		expected = "B"
	default:
		return false
	}
	pointer, ok := function.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	aliases := make(map[string]struct{})
	dotImport := false
	for _, item := range file.Imports {
		path, err := strconv.Unquote(item.Path.Value)
		if err != nil || path != "testing" {
			continue
		}
		if item.Name == nil {
			aliases["testing"] = struct{}{}
			continue
		}
		switch item.Name.Name {
		case ".":
			dotImport = true
		case "_":
		default:
			aliases[item.Name.Name] = struct{}{}
		}
	}
	if identifier, ok := pointer.X.(*ast.Ident); ok {
		return dotImport && identifier.Name == expected
	}
	selector, ok := pointer.X.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != expected {
		return false
	}
	identifier, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}
	_, ok = aliases[identifier.Name]
	return ok
}

func executableName(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	rest := name[len(prefix):]
	if rest == "" {
		return true
	}
	var runeValue rune
	for _, runeValue = range rest {
		break
	}
	return !unicode.IsLower(runeValue)
}
