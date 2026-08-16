# FEAT-002 DB schema

## 所有とmigration

FEAT-002は生成要求、試行履歴、検証済みHTML候補を所有する。migrationは既存のmigration runnerで、認証・CSRF schemaの後に順序番号`003`として一度だけ適用する。migrationはtransaction内でtable、constraint、indexを作成し、失敗時は全変更をrollbackする。

候補を合格済み版へ採用するtable、コンテンツtable、公開状態は作成しない。`source_version_id`はFEAT-003所有の安定UUIDを論理参照するため、現時点では外部FKを作成しない。修正生成の開始時に`VersionSource`が参照可能な合格済み版かを検証する。

## Table

| table | rowの意味 | 主な制約 |
| --- | --- | --- |
| `generation_requests` | 一つの初回または修正生成の最終記録 | `state`は`running`、`completed_succeeded`、`completed_failed`。`kind`に応じて初回のtopicまたは修正元versionを必須にする。 |
| `generation_attempts` | 一回の外部取得と形式検証の結果 | `(generation_request_id, attempt_number)`は一意。番号は1〜4。成功・失敗の内容整合をCHECKする。 |
| `generated_html_candidates` | 最初に形式合格した不変HTML | 生成要求と1対0..1。空HTMLを禁止し、生成後に更新・削除するFEAT-002 operationを持たない。 |

### `generation_requests`

| column | type | null | 規則 |
| --- | --- | --- | --- |
| `id` | UUID | no | applicationが生成する主キー。 |
| `kind` | TEXT | no | `initial`または`revision`。 |
| `topic` | TEXT | conditional | `initial`ではtrim後に空でない。`revision`ではNULL。 |
| `instructions` | TEXT | yes | 任意の追加指示。`revision`ではtrim後に空でない。 |
| `source_version_id` | UUID | conditional | `revision`では必須、`initial`ではNULL。FEAT-003のversion IDを論理参照する。 |
| `state` | TEXT | no | 上記の3状態のみ。 |
| `final_failure_code` | TEXT | conditional | `completed_failed`時だけ`generation_unavailable`または`invalid_html`。 |
| `final_failure_summary` | TEXT | conditional | `completed_failed`時だけ安全な要約。 |
| `created_at` | TIMESTAMPTZ | no | Server clockで記録。 |
| `completed_at` | TIMESTAMPTZ | yes | 完了状態時だけ記録。 |

`kind = initial`は`topic IS NOT NULL AND btrim(topic) <> '' AND source_version_id IS NULL`、`kind = revision`は`topic IS NULL AND source_version_id IS NOT NULL AND instructions IS NOT NULL AND btrim(instructions) <> ''`を満たす。完了成功はfailure fieldがNULL、完了失敗は両failure fieldが非NULL、`running`は`completed_at`とfailure fieldがNULLでなければならない。

### `generation_attempts`

| column | type | null | 規則 |
| --- | --- | --- | --- |
| `id` | UUID | no | applicationが生成する主キー。 |
| `generation_request_id` | UUID | no | `generation_requests(id)`へのFK。 |
| `attempt_number` | SMALLINT | no | 1〜4の連番。 |
| `outcome` | TEXT | no | `succeeded`または`failed`。 |
| `failure_code` | TEXT | conditional | failed時だけ`generation_unavailable`または`invalid_html`。 |
| `failure_summary` | TEXT | conditional | failed時だけ安全な要約。 |
| `started_at` | TIMESTAMPTZ | no | 外部呼出し前に記録。 |
| `completed_at` | TIMESTAMPTZ | no | 試行結果確定時に記録。 |

成功試行はfailure fieldを持たず、失敗試行は両failure fieldを持つ。`completed_at >= started_at`をCHECKする。request FKは`ON DELETE RESTRICT`であり、本Featureに削除operationはない。

### `generated_html_candidates`

| column | type | null | 規則 |
| --- | --- | --- | --- |
| `id` | UUID | no | FEAT-003/005へ渡す不透明で安定した候補ID。 |
| `generation_request_id` | UUID | no | `generation_requests(id)`への一意なFK。 |
| `html` | TEXT | no | 形式検証に合格した空でない完全HTMLだけ。 |
| `validated_at` | TIMESTAMPTZ | no | HTML validatorの合格時刻。 |
| `created_at` | TIMESTAMPTZ | no | 候補確定時刻。 |

候補はappend-onlyである。repositoryは候補HTMLに対するUPDATE/DELETEを発行せず、migrationでUPDATE/DELETEを拒否するtriggerも作成する。候補を返す管理HTTPはHTML本文を含めない。将来の保持・削除方針を変更する場合は、FEAT-003の採用後ライフサイクル設計と明示的に整合させ、triggerを変更するmigrationを追加する。

## DDL相当

```sql
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
```

## Access map and transaction boundary

| operation | read | write | transaction |
| --- | --- | --- | --- |
| `create_generation_request` | 修正時だけsource version、mutation guardの管理session | session idle期限、request、各attempt、成功時candidate | T1でidle期限とrequest作成、T2で途中失敗、T3で成功確定、T4で最終失敗確定を行う。 |
| `get_generation_request` | request、attempt、candidate metadata | なし | 読取り専用。 |

外部Codex呼出しの間、DB transactionを開いたままにしない。T1はFEAT-001の管理mutation guardが更新するsession idle期限とrequest作成の単一transactionである。T3の成功確定は「成功attempt、candidate、request state」の単一transaction、T4の最終失敗確定は「4回目failed attempt、request state」の単一transactionとする。途中の失敗attemptはT2で次試行前に確定し、障害後も履歴として残る。
