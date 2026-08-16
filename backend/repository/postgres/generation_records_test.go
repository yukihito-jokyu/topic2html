package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/yukihito-jokyu/topic2html/backend/domain/generation"
)

type fakeRows struct {
	values  [][]any
	index   int
	err     error
	scanErr error
}

func (r *fakeRows) Close()     {}
func (r *fakeRows) Err() error { return r.err }
func (r *fakeRows) Next() bool {
	if r.index >= len(r.values) {
		return false
	}
	r.index++

	return true
}
func (r *fakeRows) Scan(destinations ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	values := r.values[r.index-1]
	for index, destination := range destinations {
		switch value := destination.(type) {
		case *string:
			*value = values[index].(string)
		case *generation.Kind:
			*value = values[index].(generation.Kind)
		case **string:
			*value = values[index].(*string)
		case *generation.State:
			*value = values[index].(generation.State)
		case **generation.FailureCode:
			*value = values[index].(*generation.FailureCode)
		case *time.Time:
			*value = values[index].(time.Time)
		case **time.Time:
			*value = values[index].(*time.Time)
		case *int16:
			*value = values[index].(int16)
		case *generation.Outcome:
			*value = values[index].(generation.Outcome)
		default:
			return errors.New("unsupported scan destination")
		}
	}

	return nil
}

type fakeReader struct {
	queryRows []rows
	queryErr  error
	rowErrs   []error
	values    [][]any
	index     int
}

func (r *fakeReader) QueryRow(_ context.Context, _ string, _ ...any) row {
	err := pgx.ErrNoRows
	if len(r.rowErrs) > 0 {
		err = r.rowErrs[0]
		r.rowErrs = r.rowErrs[1:]
	}

	return scanValues{
		values: r.values,
		index:  &r.index,
		err:    err,
	}
}
func (r *fakeReader) Query(context.Context, string, ...any) (rows, error) {
	if r.queryErr != nil {
		return nil, r.queryErr
	}
	if len(r.queryRows) == 0 {
		return &fakeRows{}, nil
	}
	result := r.queryRows[0]
	r.queryRows = r.queryRows[1:]

	return result, nil
}

type scanValues struct {
	values [][]any
	index  *int
	err    error
}

func (r scanValues) Scan(destinations ...any) error {
	if r.err != nil {
		return r.err
	}
	values := r.values[*r.index]
	*r.index++

	return (&fakeRows{
		values: [][]any{values},
		index:  1,
	}).Scan(destinations...)
}

func generationStore(transaction *fakeTx, reader reader) *Store {
	return &Store{
		pool: fakePool{
			transaction: transaction,
		},
		reader: reader,
	}
}

func generationRequest(now time.Time) generation.Request {
	topic := "Go"

	return generation.Request{
		ID:        "00000000-0000-0000-0000-000000000001",
		Kind:      generation.KindInitial,
		Topic:     &topic,
		State:     generation.StateRunning,
		CreatedAt: now,
	}
}
func failedAttempt(now time.Time, number int16) generation.Attempt {
	code := generation.FailureInvalidHTML
	summary := "HTML文書の形式を確認できませんでした。"

	return generation.Attempt{
		ID:             "00000000-0000-0000-0000-000000000002",
		RequestID:      "00000000-0000-0000-0000-000000000001",
		Number:         number,
		Outcome:        generation.OutcomeFailed,
		FailureCode:    &code,
		FailureSummary: &summary,
		StartedAt:      now,
		CompletedAt:    now,
	}
}

func TestGenerationStoreWrites(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	request := generationRequest(now)
	transaction := &fakeTx{}
	success := failedAttempt(now, 2)
	success.Outcome = generation.OutcomeSucceeded
	success.FailureCode = nil
	success.FailureSummary = nil
	candidate := generation.Candidate{
		ID:          "00000000-0000-0000-0000-000000000003",
		RequestID:   success.RequestID,
		HTML:        "<!doctype html><html><head></head><body></body></html>",
		ValidatedAt: now,
		CreatedAt:   now,
	}
	for _, testCase := range []struct {
		name      string
		call      func() error
		wantError bool
	}{
		{
			name: "T1 creates a running request in the authorized transaction",
			call: func() error {
				return generationStore(transaction, nil).CreateRunningGenerationRequest(context.WithValue(ctx, transactionContextKey{}, transaction), request)
			},
		},
		{
			name: "T1 rejects an absent authorized transaction",
			call: func() error {
				return generationStore(&fakeTx{}, nil).CreateRunningGenerationRequest(ctx, request)
			},
			wantError: true,
		},
		{
			name: "T1 rejects an invalid request",
			call: func() error {
				return generationStore(transaction, nil).CreateRunningGenerationRequest(context.WithValue(ctx, transactionContextKey{}, transaction), generation.Request{})
			},
			wantError: true,
		},
		{
			name: "T2 records a failed attempt",
			call: func() error {
				return generationStore(&fakeTx{}, nil).RecordFailedGenerationAttempt(ctx, failedAttempt(now, 1))
			},
		},
		{
			name: "T2 rejects the fourth attempt",
			call: func() error {
				return generationStore(&fakeTx{}, nil).RecordFailedGenerationAttempt(ctx, failedAttempt(now, 4))
			},
			wantError: true,
		},
		{
			name: "T3 completes with a candidate",
			call: func() error {
				return generationStore(&fakeTx{}, nil).CompleteGenerationSucceeded(ctx, success, candidate, now)
			},
		},
		{
			name: "T4 completes after the fourth failed attempt",
			call: func() error {
				return generationStore(&fakeTx{}, nil).CompleteGenerationFailed(ctx, failedAttempt(now, 4), now)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if err := testCase.call(); (err != nil) != testCase.wantError {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestGenerationStoreWriteFailures(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	boom := errors.New("boom")
	writeFailureCases := []struct {
		name string
		call func() error
	}{
		{
			name: "final failure before fourth attempt",
			call: func() error {
				return generationStore(&fakeTx{}, nil).CompleteGenerationFailed(ctx, failedAttempt(now, 3), now)
			},
		},
		{
			name: "begin transaction",
			call: func() error {
				return (&Store{
					pool: fakePool{
						err: boom,
					},
				}).RecordFailedGenerationAttempt(ctx, failedAttempt(now, 1))
			},
		},
		{
			name: "record failed attempt",
			call: func() error {
				return generationStore(&fakeTx{
					exec: []error{boom},
				}, nil).RecordFailedGenerationAttempt(ctx, failedAttempt(now, 1))
			},
		},
		{
			name: "commit failed attempt",
			call: func() error {
				return generationStore(&fakeTx{
					commit: boom,
				}, nil).RecordFailedGenerationAttempt(ctx, failedAttempt(now, 1))
			},
		},
		{
			name: "final failure state update",
			call: func() error {
				return generationStore(&fakeTx{
					tag: fakeTag(0),
				}, nil).CompleteGenerationFailed(ctx, failedAttempt(now, 4), now)
			},
		},
	}
	for _, testCase := range writeFailureCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := testCase.call(); err == nil {
				t.Fatal("write failure succeeded")
			}
		})
	}
	success := failedAttempt(now, 1)
	success.Outcome = generation.OutcomeSucceeded
	success.FailureCode = nil
	success.FailureSummary = nil
	candidate := generation.Candidate{
		ID:          "candidate",
		RequestID:   success.RequestID,
		HTML:        "<!doctype html>",
		ValidatedAt: now,
		CreatedAt:   now,
	}
	completionFailureCases := []struct {
		name string
		call func() error
	}{
		{
			name: "successful completion with failed attempt",
			call: func() error {
				return generationStore(&fakeTx{}, nil).CompleteGenerationSucceeded(ctx, failedAttempt(now, 1), candidate, now)
			},
		},
		{
			name: "successful completion attempt insert",
			call: func() error {
				return generationStore(&fakeTx{
					exec: []error{boom},
				}, nil).CompleteGenerationSucceeded(ctx, success, candidate, now)
			},
		},
		{
			name: "successful completion candidate insert",
			call: func() error {
				return generationStore(&fakeTx{
					exec: []error{nil, boom},
				}, nil).CompleteGenerationSucceeded(ctx, success, candidate, now)
			},
		},
		{
			name: "successful completion request update",
			call: func() error {
				return generationStore(&fakeTx{
					exec: []error{nil, nil, boom},
				}, nil).CompleteGenerationSucceeded(ctx, success, candidate, now)
			},
		},
		{
			name: "successful completion request update no rows",
			call: func() error {
				return generationStore(&fakeTx{
					tags: []pgconnTag{fakeTag(1), fakeTag(1), fakeTag(0)},
				}, nil).CompleteGenerationSucceeded(ctx, success, candidate, now)
			},
		},
		{
			name: "failed completion attempt insert",
			call: func() error {
				return generationStore(&fakeTx{
					exec: []error{nil, boom},
				}, nil).CompleteGenerationFailed(ctx, failedAttempt(now, 4), now)
			},
		},
		{
			name: "failed completion request update no rows",
			call: func() error {
				return generationStore(&fakeTx{
					tags: []pgconnTag{fakeTag(1), fakeTag(0)},
				}, nil).CompleteGenerationFailed(ctx, failedAttempt(now, 4), now)
			},
		},
		{
			name: "failed completion request update",
			call: func() error {
				return generationStore(&fakeTx{
					exec: []error{nil, boom},
				}, nil).CompleteGenerationFailed(ctx, failedAttempt(now, 4), now)
			},
		},
	}
	for _, testCase := range completionFailureCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := testCase.call(); err == nil {
				t.Fatal("completion failure succeeded")
			}
		})
	}
}

func TestFindGenerationRequest(t *testing.T) {
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	topic := "Go"
	requestID := "00000000-0000-0000-0000-000000000001"
	candidateID := "00000000-0000-0000-0000-000000000003"
	successReader := &fakeReader{
		rowErrs: []error{nil, nil},
		values: [][]any{
			{requestID, generation.KindInitial, &topic, (*string)(nil), (*string)(nil), generation.StateCompletedSucceeded, (*generation.FailureCode)(nil), (*string)(nil), now, &now},
			{candidateID, requestID, now, now},
		},
		queryRows: []rows{&fakeRows{
			values: [][]any{{"00000000-0000-0000-0000-000000000002", requestID, int16(1), generation.OutcomeSucceeded, (*generation.FailureCode)(nil), (*string)(nil), now, now}},
		}},
	}
	noCandidateReader := &fakeReader{
		rowErrs:   []error{nil, pgx.ErrNoRows},
		values:    [][]any{{requestID, generation.KindInitial, &topic, (*string)(nil), (*string)(nil), generation.StateRunning, (*generation.FailureCode)(nil), (*string)(nil), now, (*time.Time)(nil)}},
		queryRows: []rows{&fakeRows{}},
	}
	for _, testCase := range []struct {
		name          string
		reader        reader
		attempts      int
		wantCandidate bool
	}{
		{
			name:          "completed request returns attempts and candidate metadata",
			reader:        successReader,
			attempts:      1,
			wantCandidate: true,
		},
		{
			name:   "running request has no candidate",
			reader: noCandidateReader,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			record, found, err := generationStore(nil, testCase.reader).FindGenerationRequest(context.Background(), requestID)
			if err != nil || !found || len(record.Attempts) != testCase.attempts || (record.Candidate != nil) != testCase.wantCandidate {
				t.Fatalf("record=%+v found=%t err=%v", record, found, err)
			}
			if record.Candidate != nil && record.Candidate.HTML != "" {
				t.Fatalf("candidate HTML was exposed: %q", record.Candidate.HTML)
			}
		})
	}
	for _, testCase := range []struct {
		name      string
		reader    reader
		found     bool
		wantError bool
	}{
		{
			name: "request is absent",
			reader: &fakeReader{
				rowErrs: []error{pgx.ErrNoRows},
			},
		},
		{
			name: "request lookup fails",
			reader: &fakeReader{
				rowErrs: []error{errors.New("boom")},
			},
			wantError: true,
		},
		{
			name: "attempt lookup fails",
			reader: &fakeReader{
				rowErrs:  []error{nil},
				values:   [][]any{{requestID, generation.KindInitial, &topic, (*string)(nil), (*string)(nil), generation.StateRunning, (*generation.FailureCode)(nil), (*string)(nil), now, (*time.Time)(nil)}},
				queryErr: errors.New("boom"),
			},
			wantError: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, found, err := generationStore(nil, testCase.reader).FindGenerationRequest(context.Background(), requestID)
			if found != testCase.found || (err != nil) != testCase.wantError {
				t.Fatalf("found=%t err=%v", found, err)
			}
		})
	}
	for _, testCase := range []struct {
		name string
		rows rows
	}{
		{
			name: "rows iteration fails",
			rows: &fakeRows{
				values: [][]any{{"a", requestID, int16(1), generation.OutcomeSucceeded, (*generation.FailureCode)(nil), (*string)(nil), now, now}},
				err:    errors.New("boom"),
			},
		},
		{
			name: "candidate lookup fails after attempts",
			rows: &fakeRows{
				values: [][]any{{"a", requestID, int16(1), generation.OutcomeSucceeded, (*generation.FailureCode)(nil), (*string)(nil), now, now}},
			},
		},
		{
			name: "attempt scan fails",
			rows: &fakeRows{
				values:  [][]any{{"a", requestID, int16(1), generation.OutcomeSucceeded, (*generation.FailureCode)(nil), (*string)(nil), now, now}},
				scanErr: errors.New("boom"),
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			reader := &fakeReader{
				rowErrs:   []error{nil, errors.New("boom")},
				values:    [][]any{{requestID, generation.KindInitial, &topic, (*string)(nil), (*string)(nil), generation.StateRunning, (*generation.FailureCode)(nil), (*string)(nil), now, (*time.Time)(nil)}},
				queryRows: []rows{testCase.rows},
			}
			if _, _, err := generationStore(nil, reader).FindGenerationRequest(context.Background(), requestID); err == nil {
				t.Fatal("row failure succeeded")
			}
		})
	}
}
