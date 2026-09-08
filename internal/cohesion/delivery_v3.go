package cohesion

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const cohesionGoalID = "golib-cohesion-v1"
const maximumManifestDeliveryBytes = 1 << 20

// DimensionState describes the lifecycle state of a delivery dimension.
type DimensionState string

const (
	// DimensionNotStarted indicates that work has not begun.
	DimensionNotStarted DimensionState = "not-started"
	// DimensionInProgress indicates that work is underway.
	DimensionInProgress DimensionState = "in-progress"
	// DimensionBlocked indicates that work cannot proceed.
	DimensionBlocked DimensionState = "blocked"
	// DimensionNotApplicable indicates that a dimension does not apply.
	DimensionNotApplicable DimensionState = "not-applicable"
	// DimensionVerified indicates that a dimension has passed its checks.
	DimensionVerified DimensionState = "verified"
)

// GoalStatus describes the aggregate delivery goal state.
type GoalStatus string

const (
	// GoalNotStarted indicates that no delivery dimension has started.
	GoalNotStarted GoalStatus = "not-started"
	// GoalInProgress indicates that at least one dimension is underway.
	GoalInProgress GoalStatus = "in-progress"
	// GoalBlocked indicates that at least one dimension is blocked.
	GoalBlocked GoalStatus = "blocked"
	// GoalComplete indicates that all required dimensions are verified.
	GoalComplete GoalStatus = "complete"
	// GoalNotApplicable indicates that no delivery dimension applies.
	GoalNotApplicable GoalStatus = "not-applicable"
)

// DeliveryGoal identifies the frozen cohesion delivery goal.
type DeliveryGoal struct {
	ID                 string     `json:"id"`
	RequirementsSHA256 string     `json:"requirements_sha256"`
	Status             GoalStatus `json:"status"`
}

// EvidenceReference identifies one delivery evidence artifact.
type EvidenceReference struct {
	ReceiptPath   string `json:"receipt_path"`
	ReceiptSHA256 string `json:"receipt_sha256"`
	EntryID       string `json:"entry_id"`
}

// ManifestEvidence contains delivery evidence references.
type ManifestEvidence struct {
	Implementation []EvidenceReference `json:"implementation"`
	Hardening      []EvidenceReference `json:"hardening"`
	Release        []EvidenceReference `json:"release"`
}

// ManifestDelivery contains the goal and evidence for a module delivery.
type ManifestDelivery struct {
	Goal           DeliveryGoal     `json:"goal"`
	Implementation DimensionState   `json:"implementation"`
	Hardening      DimensionState   `json:"hardening"`
	Release        DimensionState   `json:"release"`
	Evidence       ManifestEvidence `json:"evidence"`
}

var v3SHA256Pattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func validateManifestDeliveryJSON(data []byte, requirementsSHA256 string) error {
	budget := newTrustJSONBudget(maximumManifestDeliveryBytes)
	document, err := parseTrustJSONWithBudgetAndLimits(data, maximumManifestDeliveryBytes, budget, manifestDeliveryTrustJSONLimits())
	if err != nil {
		return fmt.Errorf("json-delivery: %w", err)
	}
	if err := compiledManifestDeliverySchema.Validate(document); err != nil {
		return fmt.Errorf("schema-delivery: %w", err)
	}
	var delivery ManifestDelivery
	if err := assignTrustJSON(&delivery, document, budget); err != nil {
		return fmt.Errorf("json-delivery: %w", err)
	}
	if err := validateManifestDelivery(delivery, requirementsSHA256); err != nil {
		return fmt.Errorf("manifest-delivery: %w", err)
	}
	return nil
}

func manifestDeliveryTrustJSONLimits() trustJSONParserLimits {
	return trustJSONParserLimits{maximumStringBytes: func(path []trustJSONPathStep) int {
		if len(path) == 0 || path[len(path)-1].kind != trustJSONObjectMember {
			return maximumTrustJSONString
		}
		switch path[len(path)-1].name {
		case "receipt_path":
			return 4096
		case "id", "entry_id":
			return 128
		default:
			return maximumTrustJSONString
		}
	}}
}

func deriveGoalStatus(implementation, hardening DimensionState) GoalStatus {
	implementationTerminal := implementation == DimensionVerified || implementation == DimensionNotApplicable
	hardeningTerminal := hardening == DimensionVerified || hardening == DimensionNotApplicable
	if implementationTerminal && hardeningTerminal {
		if implementation == DimensionNotApplicable && hardening == DimensionNotApplicable {
			return GoalNotApplicable
		}
		return GoalComplete
	}
	if implementation == DimensionBlocked || hardening == DimensionBlocked {
		return GoalBlocked
	}
	implementationUnstarted := implementation == DimensionNotStarted || implementation == DimensionNotApplicable
	hardeningUnstarted := hardening == DimensionNotStarted || hardening == DimensionNotApplicable
	if implementationUnstarted && hardeningUnstarted {
		return GoalNotStarted
	}
	return GoalInProgress
}

func evidenceKindForDimensionState(dimension string, state DimensionState) (string, bool) {
	switch state {
	case DimensionNotStarted, DimensionInProgress:
		return "", false
	case DimensionBlocked:
		return "blocked-decision", validDimension(dimension)
	case DimensionNotApplicable:
		return "not-applicable-decision", validDimension(dimension)
	case DimensionVerified:
		switch dimension {
		case "implementation":
			return "source-acceptance", true
		case "hardening":
			return "gate-attestation", true
		case "release":
			return "release-attestation", true
		}
	}
	return "", false
}

func validateManifestDelivery(delivery ManifestDelivery, requirementsSHA256 string) error {
	return validateManifestDeliverySelection(delivery, requirementsSHA256, make(map[string]string))
}

func validateManifestDeliverySelection(delivery ManifestDelivery, requirementsSHA256 string, selected map[string]string) error {
	if delivery.Goal.ID != cohesionGoalID {
		return fmt.Errorf("delivery goal ID must be %q", cohesionGoalID)
	}
	if !v3SHA256Pattern.MatchString(requirementsSHA256) || delivery.Goal.RequirementsSHA256 != requirementsSHA256 {
		return errors.New("delivery goal requirements digest does not match the frozen contract")
	}
	for dimension, state := range map[string]DimensionState{
		"implementation": delivery.Implementation,
		"hardening":      delivery.Hardening,
		"release":        delivery.Release,
	} {
		if !validDimensionState(state) {
			return fmt.Errorf("delivery %s state %q is invalid", dimension, state)
		}
	}
	if want := deriveGoalStatus(delivery.Implementation, delivery.Hardening); delivery.Goal.Status != want {
		return fmt.Errorf("delivery goal status %q does not equal derived status %q", delivery.Goal.Status, want)
	}

	for _, dimension := range []struct {
		name       string
		state      DimensionState
		references []EvidenceReference
	}{
		{"implementation", delivery.Implementation, delivery.Evidence.Implementation},
		{"hardening", delivery.Hardening, delivery.Evidence.Hardening},
		{"release", delivery.Release, delivery.Evidence.Release},
	} {
		if err := validateEvidenceReferences(dimension.name, dimension.state, dimension.references, selected); err != nil {
			return err
		}
	}
	return nil
}

func validateEvidenceReferences(dimension string, state DimensionState, references []EvidenceReference, selected map[string]string) error {
	_, terminal := evidenceKindForDimensionState(dimension, state)
	if terminal && len(references) == 0 {
		return fmt.Errorf("terminal %s delivery state requires evidence", dimension)
	}
	if !terminal && len(references) != 0 {
		return fmt.Errorf("nonterminal %s delivery state forbids evidence", dimension)
	}
	for index, reference := range references {
		if !utf8.ValidString(reference.ReceiptPath) || !safeRelativePath(reference.ReceiptPath) || len(reference.ReceiptPath) > 4096 {
			return fmt.Errorf("%s evidence reference %d has an unsafe receipt path", dimension, index)
		}
		if !v3SHA256Pattern.MatchString(reference.ReceiptSHA256) {
			return fmt.Errorf("%s evidence reference %d has an invalid receipt digest", dimension, index)
		}
		if !canonicalIdentity(reference.EntryID) {
			return fmt.Errorf("%s evidence reference %d has an invalid entry identity", dimension, index)
		}
		if index != 0 && compareEvidenceReference(references[index-1], reference) == 1 {
			return fmt.Errorf("%s evidence references are not sorted and unique", dimension)
		}
		key := reference.ReceiptPath + "\x00" + reference.ReceiptSHA256 + "\x00" + reference.EntryID
		if previous, exists := selected[key]; exists {
			return fmt.Errorf("receipt entry is selected for both %s and %s", previous, dimension)
		}
		selected[key] = dimension
	}
	return nil
}

func validDimensionState(state DimensionState) bool {
	switch state {
	case DimensionNotStarted, DimensionInProgress, DimensionBlocked, DimensionNotApplicable, DimensionVerified:
		return true
	default:
		return false
	}
}

func validDimension(dimension string) bool {
	return dimension == "implementation" || dimension == "hardening" || dimension == "release"
}

func canonicalIdentity(value string) bool {
	return value != "" && len(value) <= 128 && utf8.ValidString(value) && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\x00\r\n\t")
}

func compareEvidenceReference(left, right EvidenceReference) int {
	if value := strings.Compare(left.ReceiptPath, right.ReceiptPath); value != 0 {
		return value
	}
	if value := strings.Compare(left.ReceiptSHA256, right.ReceiptSHA256); value != 0 {
		return value
	}
	return strings.Compare(left.EntryID, right.EntryID)
}
