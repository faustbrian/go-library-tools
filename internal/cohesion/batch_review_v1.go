package cohesion

import (
	"bytes"
	"fmt"
)

type schemaV3BatchOracleExpectation struct {
	OraclePath string
	ReviewPath string
}

// verifySchemaV3BatchReview validates the cross-record invariants that JSON
// Schema cannot express: the complete ordered manifest, unique reviewers, and
// exact decision/asset bindings.
func verifySchemaV3BatchReview(data []byte, decisionSHA256 string, expected []schemaV3BatchOracleExpectation) error {
	if len(expected) != 20 {
		return fmt.Errorf("batch review expectations contain %d oracles, want 20", len(expected))
	}
	budget := newTrustJSONBudget(4 << 20)
	value, err := parseTrustJSONWithBudget(data, 4<<20, budget)
	if err != nil {
		return fmt.Errorf("decode batch review: %w", err)
	}
	canonical, err := canonicalizeTrustJSONValue(value, len(data), budget)
	if err != nil || !bytes.Equal(data, canonical) {
		return fmt.Errorf("batch review is not RFC 8785 canonical JSON")
	}
	top, err := oracleObject(value, "batch review", "schema_id", "schema_version", "decision", "contract", "oracle_manifest", "reviews", "created_at")
	version, versionErr := oracleInteger(top["schema_version"])
	if err != nil || versionErr != nil || top["schema_id"] != "urn:golib:cohesion:schema-v3-decision-review:v1" || version != 1 {
		return fmt.Errorf("batch review identity is invalid")
	}
	decision, err := oracleObject(top["decision"], "batch review decision", "path", "bytes_sha256")
	if err != nil || decision["bytes_sha256"] != decisionSHA256 {
		return fmt.Errorf("batch review decision binding is invalid")
	}
	manifest, ok := top["oracle_manifest"].([]any)
	if !ok || len(manifest) != len(expected) {
		return fmt.Errorf("batch review manifest must contain exactly twenty entries")
	}
	lastPath := ""
	for index, raw := range manifest {
		row, rowErr := oracleObject(raw, "batch review manifest row", "oracle_path", "source_revision", "oracle_sha256", "case_count", "case_roster_sha256", "expected_results_sha256", "serialized_ceiling", "review_path", "review_sha256")
		if rowErr != nil {
			return rowErr
		}
		path, pathOK := row["oracle_path"].(string)
		reviewPath, reviewOK := row["review_path"].(string)
		if !pathOK || !reviewOK || path != expected[index].OraclePath || reviewPath != expected[index].ReviewPath || (lastPath != "" && path <= lastPath) {
			return fmt.Errorf("batch review manifest ordering or locator binding is invalid at entry %d", index)
		}
		lastPath = path
		for _, field := range []string{"oracle_sha256", "case_roster_sha256", "expected_results_sha256", "review_sha256"} {
			if digest, ok := row[field].(string); !ok || !validV3Digest(digest) {
				return fmt.Errorf("batch review manifest digest %s is invalid", field)
			}
		}
		caseCount, caseErr := oracleInteger(row["case_count"])
		ceiling, ceilingErr := oracleInteger(row["serialized_ceiling"])
		if caseErr != nil || caseCount < 1 || ceilingErr != nil || ceiling < 1 {
			return fmt.Errorf("batch review manifest numeric bounds are invalid")
		}
		if revision, ok := row["source_revision"].(string); !ok || !validGitSHA(revision) {
			return fmt.Errorf("batch review manifest source revision is invalid")
		}
	}
	reviews, ok := top["reviews"].([]any)
	if !ok || len(reviews) != 3 {
		return fmt.Errorf("batch review must contain exactly three reviews")
	}
	wantDimensions := []string{"contract", "execution-adversary", "verification"}
	seen := map[string]struct{}{}
	for index, raw := range reviews {
		row, rowOK := raw.(map[string]any)
		if !rowOK {
			return fmt.Errorf("batch review row is not an object")
		}
		for _, field := range []string{"reviewer_id", "dimension", "reviewed_sha256", "evidence_sha256", "coverage", "outcome"} {
			if _, exists := row[field]; !exists {
				return fmt.Errorf("batch review row is missing %s", field)
			}
		}
		for field := range row {
			if field != "reviewer_id" && field != "dimension" && field != "reviewed_sha256" && field != "evidence_sha256" && field != "coverage" && field != "outcome" && field != "checker_source_sha256" && field != "checker_version" {
				return fmt.Errorf("batch review row has unknown field %s", field)
			}
		}
		id, idOK := row["reviewer_id"].(string)
		if !idOK || id == "" || id == "oracle-author" {
			return fmt.Errorf("batch review reviewer identity is invalid")
		}
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("batch review reviewer identities are not distinct")
		}
		seen[id] = struct{}{}
		if row["dimension"] != wantDimensions[index] || row["reviewed_sha256"] != decisionSHA256 || row["outcome"] != "accepted-no-findings" {
			return fmt.Errorf("batch review row %d is invalid", index)
		}
		if index == 2 {
			checkerDigest, digestOK := row["checker_source_sha256"].(string)
			checkerVersion, versionOK := row["checker_version"].(string)
			if !digestOK || !validV3Digest(checkerDigest) || !versionOK || checkerVersion == "" {
				return fmt.Errorf("batch verification checker binding is invalid")
			}
		}
	}
	return nil
}
