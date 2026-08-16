package generation

import (
	"errors"
	"time"
)

type Kind string

const (
	KindInitial  Kind = "initial"
	KindRevision Kind = "revision"
)

type State string

const (
	StateRunning            State = "running"
	StateCompletedSucceeded State = "completed_succeeded"
	StateCompletedFailed    State = "completed_failed"
)

type Outcome string

const (
	OutcomeSucceeded Outcome = "succeeded"
	OutcomeFailed    Outcome = "failed"
)

type FailureCode string

const (
	FailureGenerationUnavailable FailureCode = "generation_unavailable"
	FailureInvalidHTML           FailureCode = "invalid_html"
)

type Request struct {
	ID                  string
	Kind                Kind
	Topic               *string
	Instructions        *string
	SourceVersionID     *string
	State               State
	FinalFailureCode    *FailureCode
	FinalFailureSummary *string
	CreatedAt           time.Time
	CompletedAt         *time.Time
}

type Attempt struct {
	ID             string
	RequestID      string
	Number         int16
	Outcome        Outcome
	FailureCode    *FailureCode
	FailureSummary *string
	StartedAt      time.Time
	CompletedAt    time.Time
}

type Candidate struct {
	ID          string
	RequestID   string
	HTML        string
	ValidatedAt time.Time
	CreatedAt   time.Time
}

type Record struct {
	Request   Request
	Attempts  []Attempt
	Candidate *Candidate
}

func (request Request) ValidateRunning() error {
	if request.ID == "" || request.CreatedAt.IsZero() || request.State != StateRunning || request.CompletedAt != nil || request.FinalFailureCode != nil || request.FinalFailureSummary != nil {
		return errors.New("invalid running generation request")
	}
	if request.Kind == KindInitial {
		if request.Topic == nil || *request.Topic == "" || request.SourceVersionID != nil {
			return errors.New("invalid initial generation request")
		}

		return nil
	}
	if request.Kind != KindRevision || request.Topic != nil || request.SourceVersionID == nil || *request.SourceVersionID == "" || request.Instructions == nil || *request.Instructions == "" {
		return errors.New("invalid revision generation request")
	}

	return nil
}

func (attempt Attempt) Validate() error {
	if attempt.ID == "" || attempt.RequestID == "" || attempt.Number < 1 || attempt.Number > 4 || attempt.StartedAt.IsZero() || attempt.CompletedAt.IsZero() || attempt.CompletedAt.Before(attempt.StartedAt) {
		return errors.New("invalid generation attempt")
	}
	if attempt.Outcome == OutcomeSucceeded && attempt.FailureCode == nil && attempt.FailureSummary == nil {
		return nil
	}
	if attempt.Outcome == OutcomeFailed && validFailure(attempt.FailureCode, attempt.FailureSummary) {
		return nil
	}

	return errors.New("invalid generation attempt outcome")
}

func (candidate Candidate) Validate() error {
	if candidate.ID == "" || candidate.RequestID == "" || candidate.HTML == "" || candidate.ValidatedAt.IsZero() || candidate.CreatedAt.IsZero() {
		return errors.New("invalid generated HTML candidate")
	}

	return nil
}

func ValidFailure(code *FailureCode, summary *string) bool { return validFailure(code, summary) }

func validFailure(code *FailureCode, summary *string) bool {
	return code != nil && summary != nil && *summary != "" && (*code == FailureGenerationUnavailable || *code == FailureInvalidHTML)
}
