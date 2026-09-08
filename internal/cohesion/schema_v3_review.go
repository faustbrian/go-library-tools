package cohesion

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

const maximumSchemaV3DecisionReviewBytes = 1 << 20

const frozenContractReviewSHA256 = "sha256:0793cd5175f84fdbafe4b13590d1cb48cd2450bcd2b0a4189986fcdf3e78df9a"
const frozenContractFreezeSHA256 = "sha256:d7ecce05defd00a370a9adbfc11963d0265c000e724e77a1ab44f3743604e5f4"
const frozenSchemaV3DecisionReviewSHA256 = "sha256:c371d51fc8bf5b314d8c7cf51e8b96b7d4714397145cce09cc723f8370fa628d"
const frozenSchemaV3DecisionFreezeSHA256 = "sha256:a523f5972f4d75061d935e31773a7042ebf47d3dfb4541843d478dcc4992851f"

type schemaV3ContractReview struct {
	SchemaID  string                    `json:"schema_id"`
	GoalID    string                    `json:"goal_id"`
	Contract  schemaV3ReviewedContract  `json:"contract"`
	Review    schemaV3ContractReviewRow `json:"review"`
	Release   schemaV3ContractRelease   `json:"release"`
	CreatedAt string                    `json:"created_at"`
}

type schemaV3ContractReviewRow struct {
	ReviewerID     string   `json:"reviewer_id"`
	ReviewedCommit string   `json:"reviewed_commit"`
	Outcome        string   `json:"outcome"`
	Findings       []string `json:"findings"`
}

type schemaV3ContractRelease struct {
	WorkflowRunID   int    `json:"workflow_run_id"`
	WorkflowOutcome string `json:"workflow_outcome"`
	URL             string `json:"url"`
}

type schemaV3ContractFreeze struct {
	SchemaID  string                    `json:"schema_id"`
	GoalID    string                    `json:"goal_id"`
	Contract  schemaV3ReviewedContract  `json:"contract"`
	Review    schemaV3ContractFreezeRow `json:"review"`
	CreatedAt string                    `json:"created_at"`
}

type schemaV3ContractFreezeRow struct {
	RecordPath     string `json:"record_path"`
	RecordSHA256   string `json:"record_sha256"`
	ReviewerID     string `json:"reviewer_id"`
	ReviewedCommit string `json:"reviewed_commit"`
	Outcome        string `json:"outcome"`
}

type schemaV3DecisionReview struct {
	SchemaID      string                        `json:"schema_id"`
	SchemaVersion int                           `json:"schema_version"`
	Decision      schemaV3ReviewedDecision      `json:"decision"`
	Contract      schemaV3ReviewedContract      `json:"contract"`
	ReviewCount   int                           `json:"review_count"`
	Reviews       []schemaV3DecisionReviewEntry `json:"reviews"`
	CreatedAt     string                        `json:"created_at"`
}

type schemaV3ReviewedDecision struct {
	Path        string `json:"path"`
	BytesSHA256 string `json:"bytes_sha256"`
}

type schemaV3ReviewedContract struct {
	Repository   string `json:"repository"`
	Path         string `json:"path"`
	Tag          string `json:"tag"`
	TagObjectSHA string `json:"tag_object_sha"`
	PeeledCommit string `json:"peeled_commit"`
	BytesSHA256  string `json:"bytes_sha256"`
}

type schemaV3DecisionReviewEntry struct {
	ReviewerID     string `json:"reviewer_id"`
	Dimension      string `json:"dimension"`
	Outcome        string `json:"outcome"`
	ReviewedSHA256 string `json:"reviewed_sha256"`
}

type schemaV3DecisionFreeze struct {
	SchemaID       string                        `json:"schema_id"`
	SchemaVersion  int                           `json:"schema_version"`
	Decision       schemaV3ReviewedDecision      `json:"decision"`
	Contract       schemaV3ReviewedContract      `json:"contract"`
	DecisionReview schemaV3ReviewedDecision      `json:"decision_review"`
	ReviewCount    int                           `json:"review_count"`
	Reviews        []schemaV3DecisionReviewEntry `json:"reviews"`
	CreatedAt      string                        `json:"created_at"`
}

func validateSchemaV3BootstrapControls(contractReviewData, contractFreezeData, decisionReviewData, decisionFreezeData []byte) error {
	for _, control := range []struct {
		name   string
		data   []byte
		digest string
	}{
		{"contract review", contractReviewData, frozenContractReviewSHA256},
		{"contract freeze", contractFreezeData, frozenContractFreezeSHA256},
		{"schema-v3 decision review", decisionReviewData, frozenSchemaV3DecisionReviewSHA256},
		{"schema-v3 decision freeze", decisionFreezeData, frozenSchemaV3DecisionFreezeSHA256},
	} {
		if exactBytesSHA256(control.data) != control.digest {
			return fmt.Errorf("%s does not match its frozen exact-byte digest", control.name)
		}
	}
	return validateSchemaV3BootstrapControlSemantics(contractReviewData, contractFreezeData, decisionReviewData, decisionFreezeData)
}

func validateSchemaV3BootstrapControlSemantics(contractReviewData, contractFreezeData, decisionReviewData, decisionFreezeData []byte) error {
	contractReview, err := decodeSchemaV3Control[schemaV3ContractReview](contractReviewData, contractReviewV1SchemaIdentity)
	if err != nil {
		return err
	}
	contractFreeze, err := decodeSchemaV3Control[schemaV3ContractFreeze](contractFreezeData, contractFreezeV1SchemaIdentity)
	if err != nil {
		return err
	}
	decisionReview, decisionReviewDocument, err := decodeSchemaV3DecisionReview(decisionReviewData)
	if err != nil {
		return err
	}
	if err := compiledSchemaV3DecisionReviewSchema.Validate(decisionReviewDocument); err != nil {
		return fmt.Errorf("validate schema-v3 decision review schema: %w", err)
	}
	if err := validateSchemaV3DecisionReviewSemantics(decisionReview); err != nil {
		return err
	}
	decisionFreeze, err := decodeSchemaV3Control[schemaV3DecisionFreeze](decisionFreezeData, schemaV3DecisionFreezeV1SchemaIdentity)
	if err != nil {
		return err
	}

	if contractReview.Review.ReviewedCommit != contractReview.Contract.PeeledCommit {
		return errors.New("contract review does not bind the peeled contract commit")
	}
	wantReleaseURL := "https://github.com/faustbrian/go-library-tools/releases/tag/" + contractReview.Contract.Tag
	if contractReview.Release.URL != wantReleaseURL {
		return errors.New("contract review release URL does not match its tag")
	}
	if err := validWholeSecondUTC(contractReview.CreatedAt); err != nil {
		return err
	}
	if contractFreeze.Contract != contractReview.Contract || contractFreeze.GoalID != contractReview.GoalID ||
		contractFreeze.Review.RecordPath != ".ai/cohesion/phase3/contract-releases/v1.5.5/CONTRACT_REVIEW.json" ||
		contractFreeze.Review.RecordSHA256 != exactBytesSHA256(contractReviewData) ||
		contractFreeze.Review.ReviewerID != contractReview.Review.ReviewerID ||
		contractFreeze.Review.ReviewedCommit != contractReview.Review.ReviewedCommit ||
		contractFreeze.Review.Outcome != contractReview.Review.Outcome || contractFreeze.CreatedAt != contractReview.CreatedAt {
		return errors.New("contract freeze does not bind the accepted contract review")
	}
	if decisionReview.Contract != contractReview.Contract {
		return errors.New("schema-v3 decision review does not bind the frozen contract")
	}
	if decisionFreeze.Decision != decisionReview.Decision || decisionFreeze.Contract != decisionReview.Contract ||
		decisionFreeze.DecisionReview.Path != ".ai/cohesion/phase3/SCHEMA_V3_DECISION_REVIEW.json" ||
		decisionFreeze.DecisionReview.BytesSHA256 != exactBytesSHA256(decisionReviewData) ||
		decisionFreeze.ReviewCount != decisionReview.ReviewCount ||
		!slices.Equal(decisionFreeze.Reviews, decisionReview.Reviews) || decisionFreeze.CreatedAt != decisionReview.CreatedAt {
		return errors.New("schema-v3 decision freeze does not bind the accepted decision review")
	}
	return nil
}

func decodeSchemaV3Control[T any](data []byte, schemaIdentity string) (T, error) {
	var zero T
	budget := newTrustJSONBudget(maximumSchemaV3DecisionReviewBytes)
	document, err := parseTrustJSONWithBudget(data, maximumSchemaV3DecisionReviewBytes, budget)
	if err != nil {
		return zero, err
	}
	if err := compiledSchemaV3BootstrapSchemas[schemaIdentity].Validate(document); err != nil {
		return zero, fmt.Errorf("validate schema-v3 bootstrap control schema: %w", err)
	}
	var control T
	if err := assignTrustJSON(&control, document, budget); err != nil {
		return zero, err
	}
	return control, nil
}

func validWholeSecondUTC(value string) error {
	if _, err := time.Parse("2006-01-02T15:04:05Z", value); err != nil {
		return errors.New("schema-v3 bootstrap control time is not a valid whole-second UTC timestamp")
	}
	return nil
}

func validateSchemaV3DecisionReview(data []byte) error {
	review, document, err := decodeSchemaV3DecisionReview(data)
	if err != nil {
		return err
	}
	if err := compiledSchemaV3DecisionReviewSchema.Validate(document); err != nil {
		return fmt.Errorf("validate schema-v3 decision review schema: %w", err)
	}

	return validateSchemaV3DecisionReviewSemantics(review)
}

func validateSchemaV3DecisionReviewSemantics(review schemaV3DecisionReview) error {
	reviewers := make(map[string]struct{}, len(review.Reviews))
	for _, entry := range review.Reviews {
		if len(entry.ReviewerID) > 128 {
			return errors.New("schema-v3 decision review reviewer identity exceeds 128 bytes")
		}
		if _, duplicate := reviewers[entry.ReviewerID]; duplicate {
			return errors.New("schema-v3 decision review contains a duplicate reviewer")
		}
		reviewers[entry.ReviewerID] = struct{}{}
		if entry.ReviewedSHA256 != review.Decision.BytesSHA256 {
			return errors.New("schema-v3 decision review entry does not bind the reviewed decision")
		}
	}
	if err := validWholeSecondUTC(review.CreatedAt); err != nil {
		return errors.New("schema-v3 decision review creation time is not a valid whole-second UTC timestamp")
	}
	return nil
}

func decodeSchemaV3DecisionReview(data []byte) (schemaV3DecisionReview, any, error) {
	budget := newTrustJSONBudget(maximumSchemaV3DecisionReviewBytes)
	document, err := parseTrustJSONWithBudget(data, maximumSchemaV3DecisionReviewBytes, budget)
	if err != nil {
		return schemaV3DecisionReview{}, nil, err
	}
	var review schemaV3DecisionReview
	if err := assignTrustJSON(&review, document, budget); err != nil {
		return schemaV3DecisionReview{}, nil, err
	}
	return review, document, nil
}
