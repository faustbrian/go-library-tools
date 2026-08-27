package mutation

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
)

const (
	// GremlinsVersion is the exact mutation engine release used by the verifier.
	GremlinsVersion  = "v0.6.0"
	gremlinsSum      = "h1:3G2ROO0I3q4bb5bxElQIUITTuEbl1iOfVYFqunGwrJI="
	gremlinsGoModSum = "h1:LLbvJR33CWsu1sgvQ4qMzU2rqkwYJK3Qy/Al59eHKjA="
)

//go:embed assets/*
var verifierFiles embed.FS

var verifierAssetNames = []struct {
	identity string
	embedded string
}{
	{"scripts/internal/mutation-command.sh", "assets/mutation-command.sh"},
	{"scripts/internal/mutation-coverage.sh", "assets/mutation-coverage.sh"},
	{"scripts/patches/gremlins-run-all-mutants.patch", "assets/gremlins-run-all-mutants.patch"},
	{"scripts/patches/gremlins-shared-coverage.patch", "assets/gremlins-shared-coverage.patch"},
	{"scripts/patches/gremlins-module-relative-diff.patch", "assets/gremlins-module-relative-diff.patch"},
}

// VerifierAssets returns independent copies of the exact legacy verifier
// inputs retained for evidence validation and patched verifier builds.
func VerifierAssets() map[string][]byte {
	assets := make(map[string][]byte, len(verifierAssetNames))
	for _, asset := range verifierAssetNames {
		data, _ := verifierFiles.ReadFile(asset.embedded)
		assets[asset.identity] = append([]byte(nil), data...)
	}
	return assets
}

// LegacyVerifierDigest reproduces the verifier identity recorded by existing
// approved mutation checkpoints.
func LegacyVerifierDigest() string {
	assets := VerifierAssets()
	hash := sha256.New()
	_, _ = fmt.Fprintf(hash, "gremlins-version\t%s\n", GremlinsVersion)
	_, _ = fmt.Fprintf(hash, "gremlins-sum\t%s\n", gremlinsSum)
	_, _ = fmt.Fprintf(hash, "gremlins-gomod-sum\t%s\n", gremlinsGoModSum)
	for _, asset := range verifierAssetNames {
		assetHash := sha256.Sum256(assets[asset.identity])
		_, _ = fmt.Fprintf(hash, "file\t%s\t%s\n", asset.identity, hex.EncodeToString(assetHash[:]))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
