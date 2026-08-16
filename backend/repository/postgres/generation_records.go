package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/yukihito-jokyu/topic2html/backend/domain/generation"
)

// CreateRunningGenerationRequestはT1中の管理mutation transactionへrunning requestを保存します。
func (s *Store) CreateRunningGenerationRequest(ctx context.Context, request generation.Request) error {
	if err := request.ValidateRunning(); err != nil {
		return err
	}
	transaction, ok := transactionFromContext(ctx)
	if !ok {
		return errors.New("generation request requires an authorized admin state change transaction")
	}
	_, err := transaction.Exec(ctx, `INSERT INTO generation_requests (id, kind, topic, instructions, source_version_id, state, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, request.ID, request.Kind, request.Topic, request.Instructions, request.SourceVersionID, request.State, request.CreatedAt)

	return err
}

// RecordFailedGenerationAttemptはT2で1〜3回目の失敗attemptを確定します。
func (s *Store) RecordFailedGenerationAttempt(ctx context.Context, attempt generation.Attempt) error {
	if err := attempt.Validate(); err != nil || attempt.Outcome != generation.OutcomeFailed || attempt.Number == 4 {
		return errors.New("invalid intermediate failed generation attempt")
	}

	return s.withTransaction(ctx, func(transaction tx) error {
		return insertNextAttempt(ctx, transaction, attempt)
	})
}

// CompleteGenerationSucceededはT3で成功attempt、candidate、request状態を一括確定します。
func (s *Store) CompleteGenerationSucceeded(ctx context.Context, attempt generation.Attempt, candidate generation.Candidate, completedAt time.Time) error {
	if err := attempt.Validate(); err != nil || attempt.Outcome != generation.OutcomeSucceeded || candidate.Validate() != nil || candidate.RequestID != attempt.RequestID || completedAt.IsZero() {
		return errors.New("invalid successful generation completion")
	}

	return s.withTransaction(ctx, func(transaction tx) error {
		if err := insertNextAttempt(ctx, transaction, attempt); err != nil {
			return err
		}
		if _, err := transaction.Exec(ctx, `INSERT INTO generated_html_candidates (id, generation_request_id, html, validated_at, created_at) VALUES ($1,$2,$3,$4,$5)`, candidate.ID, candidate.RequestID, candidate.HTML, candidate.ValidatedAt, candidate.CreatedAt); err != nil {
			return err
		}
		tag, err := transaction.Exec(ctx, `UPDATE generation_requests SET state = $1, completed_at = $2 WHERE id = $3 AND state = $4`, generation.StateCompletedSucceeded, completedAt, attempt.RequestID, generation.StateRunning)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return errors.New("running generation request was not found")
		}

		return nil
	})
}

// CompleteGenerationFailedはT4で4回目failed attemptとrequest状態を一括確定します。
func (s *Store) CompleteGenerationFailed(ctx context.Context, attempt generation.Attempt, completedAt time.Time) error {
	if err := attempt.Validate(); err != nil || attempt.Outcome != generation.OutcomeFailed || attempt.Number != 4 || !generation.ValidFailure(attempt.FailureCode, attempt.FailureSummary) || completedAt.IsZero() {
		return errors.New("invalid failed generation completion")
	}

	return s.withTransaction(ctx, func(transaction tx) error {
		if err := insertNextAttempt(ctx, transaction, attempt); err != nil {
			return err
		}
		tag, err := transaction.Exec(ctx, `UPDATE generation_requests SET state = $1, completed_at = $2, final_failure_code = $3, final_failure_summary = $4 WHERE id = $5 AND state = $6`, generation.StateCompletedFailed, completedAt, *attempt.FailureCode, *attempt.FailureSummary, attempt.RequestID, generation.StateRunning)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return errors.New("running generation request was not found")
		}

		return nil
	})
}

// FindGenerationRequestは候補本文を含めずにrequest、attempt、candidate metadataをread-onlyで取得します。
func (s *Store) FindGenerationRequest(ctx context.Context, id string) (generation.Record, bool, error) {
	var record generation.Record
	err := s.reader.QueryRow(ctx, `SELECT id, kind, topic, instructions, source_version_id, state, final_failure_code, final_failure_summary, created_at, completed_at FROM generation_requests WHERE id = $1`, id).Scan(&record.Request.ID, &record.Request.Kind, &record.Request.Topic, &record.Request.Instructions, &record.Request.SourceVersionID, &record.Request.State, &record.Request.FinalFailureCode, &record.Request.FinalFailureSummary, &record.Request.CreatedAt, &record.Request.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return generation.Record{}, false, nil
	}
	if err != nil {
		return generation.Record{}, false, err
	}
	rows, err := s.reader.Query(ctx, `SELECT id, generation_request_id, attempt_number, outcome, failure_code, failure_summary, started_at, completed_at FROM generation_attempts WHERE generation_request_id = $1 ORDER BY attempt_number`, id)
	if err != nil {
		return generation.Record{}, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var attempt generation.Attempt
		if err := rows.Scan(&attempt.ID, &attempt.RequestID, &attempt.Number, &attempt.Outcome, &attempt.FailureCode, &attempt.FailureSummary, &attempt.StartedAt, &attempt.CompletedAt); err != nil {
			return generation.Record{}, false, err
		}
		record.Attempts = append(record.Attempts, attempt)
	}
	if err := rows.Err(); err != nil {
		return generation.Record{}, false, err
	}
	var candidate generation.Candidate
	err = s.reader.QueryRow(ctx, `SELECT id, generation_request_id, validated_at, created_at FROM generated_html_candidates WHERE generation_request_id = $1`, id).Scan(&candidate.ID, &candidate.RequestID, &candidate.ValidatedAt, &candidate.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return record, true, nil
	}
	if err != nil {
		return generation.Record{}, false, err
	}
	record.Candidate = &candidate

	return record, true, nil
}

func (s *Store) withTransaction(ctx context.Context, operation func(tx) error) error {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if err := operation(transaction); err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

func insertNextAttempt(ctx context.Context, transaction tx, attempt generation.Attempt) error {
	tag, err := transaction.Exec(ctx, `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, failure_code, failure_summary, started_at, completed_at) SELECT $1,$2,$3,$4,$5,$6,$7,$8 WHERE EXISTS (SELECT 1 FROM generation_requests WHERE id = $2 AND state = $9) AND $3::SMALLINT = COALESCE((SELECT MAX(attempt_number) + 1 FROM generation_attempts WHERE generation_request_id = $2), 1)::SMALLINT`, attempt.ID, attempt.RequestID, attempt.Number, attempt.Outcome, attempt.FailureCode, attempt.FailureSummary, attempt.StartedAt, attempt.CompletedAt, generation.StateRunning)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("generation attempt is not next for a running request")
	}

	return nil
}
