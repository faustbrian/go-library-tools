package cohesion

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

const (
	frozenSchemaV3DecisionSHA256     = "sha256:81f90106d873e8635f4aa4a8d120d5a074fdc0eb2146c767635e9aa9879d5776"
	maximumForwardOracleFixtureBytes = 4 << 20
	maximumForwardOracleSplices      = 32
	maximumForwardOracleDeleteBytes  = 65535
	maximumForwardOracleInsertBytes  = 16384
)

type forwardOracleV2Expectation struct {
	SchemaIdentity string
	CaseCount      int
	MaximumBytes   int64
	CaseIDs        []string
	OraclePath     string
}

type forwardOracleV2 struct {
	Cases []forwardOracleV2Case
}

type forwardOracleV2Case struct {
	CaseID                string
	Input                 []byte
	Outcome               string
	NormalizedValueSHA256 *string
	ErrorCode             *schemaV3ErrorCode
	inputKind             string
	fixtureID             string
}

type forwardOracleV2Splice struct {
	offset      int
	deleteCount int
	insert      []byte
}

func verifyForwardOracleV2(data []byte, expectation forwardOracleV2Expectation) error {
	oracle, err := decodeForwardOracleV2(data, expectation.MaximumBytes)
	if err != nil {
		return err
	}
	if len(oracle.Cases) != expectation.CaseCount {
		return fmt.Errorf("forward oracle case count = %d, want %d", len(oracle.Cases), expectation.CaseCount)
	}
	caseIDs := make([]string, len(oracle.Cases))
	for index, testCase := range oracle.Cases {
		caseIDs[index] = testCase.CaseID
	}
	if !slices.Equal(caseIDs, expectation.CaseIDs) {
		return fmt.Errorf("forward oracle case roster does not match frozen roster")
	}
	for _, testCase := range oracle.Cases {
		normalized, normalizeErr := validateAndNormalizeSchemaV3(expectation.SchemaIdentity, testCase.Input)
		switch testCase.Outcome {
		case "accepted":
			if normalizeErr != nil {
				return fmt.Errorf("forward oracle case %s: expected acceptance: %w", testCase.CaseID, normalizeErr)
			}
			if testCase.NormalizedValueSHA256 == nil || *testCase.NormalizedValueSHA256 != normalized.NormalizedSHA256 || testCase.ErrorCode != nil {
				return fmt.Errorf("forward oracle case %s: normalized digest or nullable outcome mismatch", testCase.CaseID)
			}
		case "rejected":
			failure, ok := normalizeErr.(schemaV3Failure)
			if !ok || testCase.ErrorCode == nil || failure.Code != *testCase.ErrorCode || testCase.NormalizedValueSHA256 != nil {
				return fmt.Errorf("forward oracle case %s: rejection mismatch", testCase.CaseID)
			}
		default:
			return fmt.Errorf("forward oracle case %s: invalid outcome %q", testCase.CaseID, testCase.Outcome)
		}
	}
	return nil
}

func verifyForwardOracleV2Review(oracleData, reviewData []byte, expectation forwardOracleV2Expectation) error {
	if expectation.MaximumBytes < 0 || int64(len(oracleData)) > expectation.MaximumBytes {
		return fmt.Errorf("forward oracle exceeds its reviewed serialized ceiling")
	}
	const maximumReviewBytes = 1 << 20
	budget := newTrustJSONBudget(maximumReviewBytes)
	value, err := parseTrustJSONWithBudget(reviewData, maximumReviewBytes, budget)
	if err != nil {
		return fmt.Errorf("decode forward-oracle review: %w", err)
	}
	canonical, err := canonicalizeTrustJSONValue(value, len(reviewData), budget)
	if err != nil {
		return fmt.Errorf("canonicalize forward-oracle review: %w", err)
	}
	if !bytes.Equal(reviewData, canonical) {
		return fmt.Errorf("forward-oracle review is not RFC 8785 canonical JSON")
	}
	top, err := oracleObject(value, "forward-oracle review", "decision_sha256", "oracle", "serialized_bytes", "serialized_ceiling", "review_count", "reviews", "outcome")
	if err != nil {
		return err
	}
	if top["decision_sha256"] != frozenSchemaV3DecisionSHA256 || top["outcome"] != "accepted-no-findings" {
		return fmt.Errorf("forward-oracle review decision or outcome is invalid")
	}
	serializedBytes, bytesErr := oracleInteger(top["serialized_bytes"])
	serializedCeiling, ceilingErr := oracleInteger(top["serialized_ceiling"])
	if bytesErr != nil || ceilingErr != nil || serializedBytes != len(oracleData) || int64(serializedCeiling) != expectation.MaximumBytes {
		return fmt.Errorf("forward-oracle review serialized byte binding is invalid")
	}
	oracle, err := oracleObject(top["oracle"], "forward-oracle review oracle", "path", "source_revision", "bytes_sha256")
	if err != nil {
		return err
	}
	sourceRevision, revisionOK := oracle["source_revision"].(string)
	if oracle["path"] != expectation.OraclePath || !revisionOK || !validGitSHA(sourceRevision) || oracle["bytes_sha256"] != exactBytesSHA256(oracleData) {
		return fmt.Errorf("forward-oracle review oracle binding is invalid")
	}
	reviewCount, countErr := oracleInteger(top["review_count"])
	reviews, reviewsOK := top["reviews"].([]any)
	if countErr != nil || !reviewsOK || reviewCount != 3 || len(reviews) != 3 {
		return fmt.Errorf("forward-oracle review must contain exactly three reviews")
	}
	dimensions := []string{"contract", "execution-adversary", "verification"}
	reviewers := make(map[string]struct{}, len(reviews))
	for index, raw := range reviews {
		review, err := oracleObject(raw, "forward-oracle review row", "reviewer_id", "dimension", "reviewed_sha256", "outcome")
		if err != nil {
			return err
		}
		reviewerID, ok := review["reviewer_id"].(string)
		if !ok || reviewerID == "" || len(reviewerID) > 128 {
			return fmt.Errorf("forward-oracle reviewer identity is invalid")
		}
		if _, duplicate := reviewers[reviewerID]; duplicate {
			return fmt.Errorf("forward-oracle reviewer identities are not distinct")
		}
		reviewers[reviewerID] = struct{}{}
		if review["dimension"] != dimensions[index] || review["reviewed_sha256"] != oracle["bytes_sha256"] || review["outcome"] != "accepted-no-findings" {
			return fmt.Errorf("forward-oracle review row %d is invalid", index)
		}
	}
	return nil
}

func validGitSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, current := range value {
		if !((current >= '0' && current <= '9') || (current >= 'a' && current <= 'f')) {
			return false
		}
	}
	return true
}

func decodeForwardOracleV2(data []byte, maximumBytes int64) (forwardOracleV2, error) {
	if maximumBytes < 0 || int64(len(data)) > maximumBytes {
		return forwardOracleV2{}, fmt.Errorf("forward oracle exceeds its maximum byte length")
	}
	budget := newTrustJSONBudget(maximumBytes)
	value, err := parseTrustJSONWithBudgetAndLimits(data, maximumBytes, budget, forwardOracleTrustJSONLimits())
	if err != nil {
		return forwardOracleV2{}, err
	}
	canonical, err := canonicalizeTrustJSONValue(value, len(data), budget)
	if err != nil {
		return forwardOracleV2{}, err
	}
	if !bytes.Equal(data, canonical) {
		return forwardOracleV2{}, fmt.Errorf("forward oracle is not RFC 8785 canonical JSON")
	}
	top, err := oracleObject(value, "forward oracle", "format", "fixture_count", "fixtures", "case_count", "cases")
	if err != nil {
		return forwardOracleV2{}, err
	}
	if top["format"] != "golib-forward-oracle-v2" {
		return forwardOracleV2{}, fmt.Errorf("forward oracle format is not golib-forward-oracle-v2")
	}
	fixtureCount, err := oracleInteger(top["fixture_count"])
	if err != nil || fixtureCount != 2 {
		return forwardOracleV2{}, fmt.Errorf("forward oracle fixture_count must equal 2")
	}
	fixtureValues, ok := top["fixtures"].([]any)
	if !ok || len(fixtureValues) != 2 {
		return forwardOracleV2{}, fmt.Errorf("forward oracle fixtures must contain exactly two rows")
	}
	wantFixtureIDs := []string{"base.canonical-minimum", "base.canonical-rich"}
	fixtures := make(map[string][]byte, 2)
	for index, raw := range fixtureValues {
		fixture, err := oracleObject(raw, "forward oracle fixture", "fixture_id", "bytes_base64", "bytes_sha256")
		if err != nil {
			return forwardOracleV2{}, err
		}
		fixtureID, ok := fixture["fixture_id"].(string)
		if !ok || fixtureID != wantFixtureIDs[index] {
			return forwardOracleV2{}, fmt.Errorf("forward oracle fixture IDs are not the frozen ordered pair")
		}
		decoded, err := decodeCanonicalBase64(fixture["bytes_base64"], maximumForwardOracleFixtureBytes)
		if err != nil {
			return forwardOracleV2{}, fmt.Errorf("forward oracle fixture %s: %w", fixtureID, err)
		}
		digest, ok := fixture["bytes_sha256"].(string)
		if !ok || digest != exactBytesSHA256(decoded) {
			return forwardOracleV2{}, fmt.Errorf("forward oracle fixture %s digest mismatch", fixtureID)
		}
		fixtures[fixtureID] = decoded
	}
	caseCount, err := oracleInteger(top["case_count"])
	if err != nil {
		return forwardOracleV2{}, fmt.Errorf("forward oracle case_count is invalid")
	}
	caseValues, ok := top["cases"].([]any)
	if !ok || caseCount != len(caseValues) {
		return forwardOracleV2{}, fmt.Errorf("forward oracle case_count does not match cases")
	}
	oracle := forwardOracleV2{Cases: make([]forwardOracleV2Case, 0, len(caseValues))}
	previousID := ""
	fixtureReferences := make(map[string]bool, len(fixtures))
	for _, raw := range caseValues {
		testCase, err := decodeForwardOracleV2Case(raw, fixtures)
		if err != nil {
			return forwardOracleV2{}, err
		}
		if previousID != "" && testCase.CaseID <= previousID {
			return forwardOracleV2{}, fmt.Errorf("forward oracle cases are not sorted uniquely by case_id")
		}
		previousID = testCase.CaseID
		fixtureReferences[testCase.fixtureID] = true
		oracle.Cases = append(oracle.Cases, testCase)
	}
	for fixtureID := range fixtures {
		if !fixtureReferences[fixtureID] {
			return forwardOracleV2{}, fmt.Errorf("forward oracle fixture %s is unreferenced", fixtureID)
		}
	}
	return oracle, nil
}

func decodeForwardOracleV2Case(raw any, fixtures map[string][]byte) (forwardOracleV2Case, error) {
	object, err := oracleObject(raw, "forward oracle case", "case_id", "input", "outcome", "normalized_value_sha256", "error_code")
	if err != nil {
		return forwardOracleV2Case{}, err
	}
	caseID, ok := object["case_id"].(string)
	if !ok || caseID == "" || len(caseID) > 128 {
		return forwardOracleV2Case{}, fmt.Errorf("forward oracle case_id is invalid")
	}
	inputObject, ok := object["input"].(map[string]any)
	if !ok {
		return forwardOracleV2Case{}, fmt.Errorf("forward oracle case %s input is invalid", caseID)
	}
	inputKind, _ := inputObject["kind"].(string)
	fixtureID, _ := inputObject["fixture_id"].(string)
	input, err := reconstructForwardOracleV2Input(inputObject, fixtures)
	if err != nil {
		return forwardOracleV2Case{}, fmt.Errorf("forward oracle case %s: %w", caseID, err)
	}
	outcome, ok := object["outcome"].(string)
	if !ok || (outcome != "accepted" && outcome != "rejected") {
		return forwardOracleV2Case{}, fmt.Errorf("forward oracle case %s outcome is invalid", caseID)
	}
	if caseID == "base.canonical-minimum" {
		if inputKind != "fixture" || fixtureID != caseID {
			return forwardOracleV2Case{}, fmt.Errorf("forward oracle minimum base case does not use its fixture")
		}
	} else if caseID == "base.canonical-rich" {
		if inputKind != "fixture" || fixtureID != caseID {
			return forwardOracleV2Case{}, fmt.Errorf("forward oracle rich base case does not use its fixture")
		}
	} else if inputKind != "splice" || fixtureID != "base.canonical-rich" {
		return forwardOracleV2Case{}, fmt.Errorf("forward oracle non-base case %s is not a rich-fixture splice", caseID)
	}
	result := forwardOracleV2Case{CaseID: caseID, Input: input, Outcome: outcome, inputKind: inputKind, fixtureID: fixtureID}
	if rawDigest := object["normalized_value_sha256"]; rawDigest != nil {
		digest, ok := rawDigest.(string)
		if !ok || !validV3Digest(digest) {
			return forwardOracleV2Case{}, fmt.Errorf("forward oracle case %s normalized digest is invalid", caseID)
		}
		result.NormalizedValueSHA256 = &digest
	}
	if rawCode := object["error_code"]; rawCode != nil {
		text, ok := rawCode.(string)
		code := schemaV3ErrorCode(text)
		if !ok {
			return forwardOracleV2Case{}, fmt.Errorf("forward oracle case %s error code is invalid", caseID)
		}
		if _, exists := schemaV3KnownErrorCodes[code]; !exists {
			return forwardOracleV2Case{}, fmt.Errorf("forward oracle case %s error code is not closed", caseID)
		}
		result.ErrorCode = &code
	}
	if (outcome == "accepted") != (result.NormalizedValueSHA256 != nil && result.ErrorCode == nil) ||
		(outcome == "rejected") != (result.NormalizedValueSHA256 == nil && result.ErrorCode != nil) {
		return forwardOracleV2Case{}, fmt.Errorf("forward oracle case %s has invalid nullable outcome fields", caseID)
	}
	return result, nil
}

func reconstructForwardOracleV2Input(raw any, fixtures map[string][]byte) ([]byte, error) {
	object, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("input is not an object")
	}
	kindValue, ok := object["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("input kind is invalid")
	}
	fixtureID, ok := object["fixture_id"].(string)
	fixture, exists := fixtures[fixtureID]
	if !ok || !exists {
		return nil, fmt.Errorf("input fixture_id is unresolved")
	}
	if kindValue == "fixture" {
		if _, err := oracleObject(object, "fixture input", "kind", "fixture_id"); err != nil {
			return nil, err
		}
		return slices.Clone(fixture), nil
	}
	if kindValue != "splice" {
		return nil, fmt.Errorf("input kind is invalid")
	}
	if _, err := oracleObject(object, "splice input", "kind", "fixture_id", "splices"); err != nil {
		return nil, err
	}
	rawSplices, ok := object["splices"].([]any)
	if !ok || len(rawSplices) > maximumForwardOracleSplices {
		return nil, fmt.Errorf("splices are invalid")
	}
	splices := make([]forwardOracleV2Splice, 0, len(rawSplices))
	totalInsert := 0
	previousEnd := 0
	for index, rawSplice := range rawSplices {
		object, err := oracleObject(rawSplice, "splice", "offset", "delete_count", "insert_base64")
		if err != nil {
			return nil, err
		}
		offset, offsetErr := oracleInteger(object["offset"])
		deleteCount, deleteErr := oracleInteger(object["delete_count"])
		insert, insertErr := decodeCanonicalBase64(object["insert_base64"], maximumForwardOracleInsertBytes)
		if offsetErr != nil || deleteErr != nil || insertErr != nil || offset < 0 || deleteCount < 0 || deleteCount > maximumForwardOracleDeleteBytes || offset+deleteCount > len(fixture) {
			return nil, fmt.Errorf("splice descriptor is out of bounds")
		}
		if index > 0 && offset < previousEnd {
			return nil, fmt.Errorf("splices overlap or are not strictly ordered")
		}
		if index > 0 && offset == splices[index-1].offset {
			return nil, fmt.Errorf("splices have duplicate offsets")
		}
		if deleteCount == 0 && len(insert) == 0 {
			return nil, fmt.Errorf("splice is an empty no-op")
		}
		totalInsert += len(insert)
		if totalInsert > maximumForwardOracleInsertBytes {
			return nil, fmt.Errorf("splice insertions exceed their byte limit")
		}
		splices = append(splices, forwardOracleV2Splice{offset: offset, deleteCount: deleteCount, insert: insert})
		previousEnd = offset + deleteCount
	}
	capacity := len(fixture) + totalInsert
	for _, splice := range splices {
		capacity -= splice.deleteCount
	}
	result := make([]byte, 0, capacity)
	cursor := 0
	for _, splice := range splices {
		result = append(result, fixture[cursor:splice.offset]...)
		result = append(result, splice.insert...)
		cursor = splice.offset + splice.deleteCount
	}
	result = append(result, fixture[cursor:]...)
	return result, nil
}

func oracleObject(raw any, label string, names ...string) (map[string]any, error) {
	object, ok := raw.(map[string]any)
	if !ok || len(object) != len(names) {
		return nil, fmt.Errorf("%s is not a closed object", label)
	}
	for _, name := range names {
		if _, exists := object[name]; !exists {
			return nil, fmt.Errorf("%s is missing %s", label, name)
		}
	}
	return object, nil
}

func oracleInteger(raw any) (int, error) {
	number, ok := raw.(json.Number)
	if !ok {
		return 0, fmt.Errorf("value is not an integer")
	}
	integer, err := number.Int64()
	if err != nil || integer < 0 || int64(int(integer)) != integer {
		return 0, fmt.Errorf("value is not a nonnegative machine integer")
	}
	return int(integer), nil
}

func decodeCanonicalBase64(raw any, maximum int) ([]byte, error) {
	text, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("base64 value is not a string")
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(text)
	if err != nil || base64.StdEncoding.EncodeToString(decoded) != text || len(decoded) > maximum {
		return nil, fmt.Errorf("base64 value is not canonical padded RFC 4648 within its limit")
	}
	return decoded, nil
}

func validV3Digest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, current := range value[len("sha256:"):] {
		if !((current >= '0' && current <= '9') || (current >= 'a' && current <= 'f')) {
			return false
		}
	}
	return true
}
