package generation

import (
	"testing"
	"time"
)

func TestRequestValidateRunning(t *testing.T) {
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	topic := "topic"
	instructions := "instructions"
	source := "00000000-0000-0000-0000-000000000002"
	for _, tt := range []struct {
		name    string
		request Request
		valid   bool
	}{
		{
			"initial",
			Request{
				ID:        "00000000-0000-0000-0000-000000000001",
				Kind:      KindInitial,
				Topic:     &topic,
				State:     StateRunning,
				CreatedAt: now,
			},
			true,
		},
		{
			"revision",
			Request{
				ID:              "00000000-0000-0000-0000-000000000001",
				Kind:            KindRevision,
				Instructions:    &instructions,
				SourceVersionID: &source,
				State:           StateRunning,
				CreatedAt:       now,
			},
			true,
		},
		{
			"completed",
			Request{
				ID:        "00000000-0000-0000-0000-000000000001",
				Kind:      KindInitial,
				Topic:     &topic,
				State:     StateCompletedSucceeded,
				CreatedAt: now,
			},
			false,
		},
		{
			"initial source",
			Request{
				ID:              "00000000-0000-0000-0000-000000000001",
				Kind:            KindInitial,
				Topic:           &topic,
				SourceVersionID: &source,
				State:           StateRunning,
				CreatedAt:       now,
			},
			false,
		},
		{
			"revision topic",
			Request{
				ID:              "00000000-0000-0000-0000-000000000001",
				Kind:            KindRevision,
				Topic:           &topic,
				Instructions:    &instructions,
				SourceVersionID: &source,
				State:           StateRunning,
				CreatedAt:       now,
			},
			false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.request.ValidateRunning() == nil; got != tt.valid {
				t.Fatalf("valid=%t, want %t", got, tt.valid)
			}
		})
	}
}

func TestAttemptAndCandidateValidate(t *testing.T) {
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	code := FailureInvalidHTML
	invalidCode := FailureCode("other")
	summary := "HTML文書の形式を確認できませんでした。"
	for _, tt := range []struct {
		name  string
		value Attempt
		valid bool
	}{
		{
			"success",
			Attempt{
				ID:          "a",
				RequestID:   "r",
				Number:      1,
				Outcome:     OutcomeSucceeded,
				StartedAt:   now,
				CompletedAt: now,
			},
			true,
		},
		{
			"failure",
			Attempt{
				ID:             "a",
				RequestID:      "r",
				Number:         4,
				Outcome:        OutcomeFailed,
				FailureCode:    &code,
				FailureSummary: &summary,
				StartedAt:      now,
				CompletedAt:    now,
			},
			true,
		},
		{
			"invalid code",
			Attempt{
				ID:             "a",
				RequestID:      "r",
				Number:         1,
				Outcome:        OutcomeFailed,
				FailureCode:    &invalidCode,
				FailureSummary: &summary,
				StartedAt:      now,
				CompletedAt:    now,
			},
			false,
		},
		{
			"backwards time",
			Attempt{
				ID:          "a",
				RequestID:   "r",
				Number:      1,
				Outcome:     OutcomeSucceeded,
				StartedAt:   now,
				CompletedAt: now.Add(-time.Second),
			},
			false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.Validate() == nil; got != tt.valid {
				t.Fatalf("valid=%t, want %t", got, tt.valid)
			}
		})
	}
	for _, tt := range []struct {
		name  string
		value Candidate
		valid bool
	}{
		{
			name: "valid candidate",
			value: Candidate{
				ID:          "c",
				RequestID:   "r",
				HTML:        "<!doctype html>",
				ValidatedAt: now,
				CreatedAt:   now,
			},
			valid: true,
		},
		{
			name: "empty candidate",
			value: Candidate{
				ID:          "c",
				RequestID:   "r",
				ValidatedAt: now,
				CreatedAt:   now,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.Validate() == nil; got != tt.valid {
				t.Fatalf("valid=%t, want %t", got, tt.valid)
			}
		})
	}
	for _, tt := range []struct {
		name    string
		code    *FailureCode
		summary *string
		valid   bool
	}{
		{
			name:    "complete failure",
			code:    &code,
			summary: &summary,
			valid:   true,
		},
		{
			name:    "missing code",
			summary: &summary,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidFailure(tt.code, tt.summary); got != tt.valid {
				t.Fatalf("valid=%t, want %t", got, tt.valid)
			}
		})
	}
}
