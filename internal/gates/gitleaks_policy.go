package gates

import _ "embed"

// gitleaksPolicy is centrally reviewed data, not consumer-controlled policy.
// Embed it so installed verifiers retain the same private configuration owner.
// The asset is deliberately not named gitleaks.toml: the inherited scanner
// defaults exclude paths containing that name from source scanning.
//
//go:embed policy/secrets.toml
var gitleaksPolicy string
