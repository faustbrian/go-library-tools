package services_test

import (
	"bytes"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/services"
)

func FuzzServiceLock(f *testing.F) {
	f.Add([]byte("opensearch_image_repository='opensearchproject/opensearch'\nopensearch_old_version='1.0.0'\nopensearch_old_digest='sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'\nopensearch_new_version='2.0.0'\nopensearch_new_digest='sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb'\n"))
	f.Add([]byte("unknown='value'\n"))
	f.Fuzz(func(_ *testing.T, input []byte) {
		if len(input) > 32<<10 {
			return
		}
		_, _ = services.ParseOpenSearchImages(bytes.NewReader(input))
	})
}
