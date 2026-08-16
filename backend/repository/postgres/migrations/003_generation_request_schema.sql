CREATE TABLE generation_requests (
  id UUID PRIMARY KEY,
  kind TEXT NOT NULL CHECK (kind IN ('initial', 'revision')),
  topic TEXT,
  instructions TEXT,
  source_version_id UUID,
  state TEXT NOT NULL CHECK (state IN ('running', 'completed_succeeded', 'completed_failed')),
  final_failure_code TEXT CHECK (final_failure_code IN ('generation_unavailable', 'invalid_html')),
  final_failure_summary TEXT,
  created_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ,
  CHECK (
    (kind = 'initial' AND topic IS NOT NULL AND btrim(topic) <> '' AND source_version_id IS NULL)
    OR (kind = 'revision' AND topic IS NULL AND source_version_id IS NOT NULL
      AND instructions IS NOT NULL AND btrim(instructions) <> '')
  ),
  CHECK (
    (state = 'running' AND completed_at IS NULL AND final_failure_code IS NULL AND final_failure_summary IS NULL)
    OR (state = 'completed_succeeded' AND completed_at IS NOT NULL
      AND final_failure_code IS NULL AND final_failure_summary IS NULL)
    OR (state = 'completed_failed' AND completed_at IS NOT NULL
      AND final_failure_code IS NOT NULL AND final_failure_summary IS NOT NULL)
  )
);

CREATE TABLE generation_attempts (
  id UUID PRIMARY KEY,
  generation_request_id UUID NOT NULL REFERENCES generation_requests(id) ON DELETE RESTRICT,
  attempt_number SMALLINT NOT NULL CHECK (attempt_number BETWEEN 1 AND 4),
  outcome TEXT NOT NULL CHECK (outcome IN ('succeeded', 'failed')),
  failure_code TEXT CHECK (failure_code IN ('generation_unavailable', 'invalid_html')),
  failure_summary TEXT,
  started_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ NOT NULL CHECK (completed_at >= started_at),
  UNIQUE (generation_request_id, attempt_number),
  CHECK (
    (outcome = 'succeeded' AND failure_code IS NULL AND failure_summary IS NULL)
    OR (outcome = 'failed' AND failure_code IS NOT NULL AND failure_summary IS NOT NULL)
  )
);

CREATE TABLE generated_html_candidates (
  id UUID PRIMARY KEY,
  generation_request_id UUID NOT NULL UNIQUE REFERENCES generation_requests(id) ON DELETE RESTRICT,
  html TEXT NOT NULL CHECK (html <> ''),
  validated_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE FUNCTION reject_generated_html_candidate_mutation()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  RAISE EXCEPTION 'generated HTML candidates are immutable';
END;
$$;

CREATE TRIGGER generated_html_candidates_immutable
BEFORE UPDATE OR DELETE ON generated_html_candidates
FOR EACH ROW EXECUTE FUNCTION reject_generated_html_candidate_mutation();

CREATE INDEX generation_attempts_request_number_idx
  ON generation_attempts (generation_request_id, attempt_number);
CREATE INDEX generation_requests_source_created_idx
  ON generation_requests (source_version_id, created_at DESC);
