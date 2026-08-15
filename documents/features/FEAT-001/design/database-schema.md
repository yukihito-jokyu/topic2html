# 管理認証のPostgreSQL schema・migration契約

## Migration

Migration runnerだけが適用する。migration versionは昇順で適用し、`001_admin_auth_schema`は先行migrationを持たない。runnerは各migration transactionで、以下のmetadata tableを作成または参照する。

```sql
-- migration適用記録を保持する。migration runnerだけが読書きする。
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT NOT NULL PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL
);
```

`schema_migrations.version`はmigration識別子（このFeatureでは`001_admin_auth_schema`）であり、同じversionの再適用を防ぐ。一つのmigrationの適用順は次のとおりとする。

1. transactionを開始する。
2. 上記DDLでmetadata tableを作成または参照し、対象versionの適用記録を確認する。記録済みならmigration DDLを実行せずcommitする。
3. 未記録なら、対象migrationのtable、制約、indexをすべて作成する。
4. すべて成功した後にだけ、次のINSERTで`schema_migrations`へ対象versionと適用時刻を記録し、同じtransactionをcommitする。

```sql
-- 対象migrationのDDLが全て成功した後にだけ適用済みとして記録する。
INSERT INTO schema_migrations (version, applied_at)
VALUES ('001_admin_auth_schema', CURRENT_TIMESTAMP);
```

失敗時はmetadata tableの初回作成、対象DDL、適用記録を含めて全てrollbackし、適用記録を残さない。中断後は未記録versionを最初から再実行する。`001_admin_auth_schema`はこの順でmetadata table、以下の二表、制約、index、適用記録を一つのDB transactionで作成する。

期限切れ・無効化済み行の削除は認証と別の保守操作とし、期限後24時間を過ぎたOAuth transaction、失効または絶対期限後24時間を過ぎたsessionだけを対象にする。削除失敗は認証を許可する理由にしない。

## `admin_oauth_transactions`

| column | 型・NULL | 制約・用途 |
| --- | --- | --- |
| `id` | UUID, NOT NULL | primary key。Server内部識別子。 |
| `reference_hash` | BYTEA, NOT NULL | cookie参照値のSHA-256。UNIQUE。 |
| `state_hash` | BYTEA, NOT NULL | callbackの`state`のSHA-256。UNIQUE。 |
| `nonce_hash` | BYTEA, NOT NULL | ID Tokenの`nonce` claimのSHA-256。 |
| `pkce_verifier_ciphertext` | BYTEA, NOT NULL | Server保護鍵で暗号化したverifier。 |
| `return_path` | TEXT, NOT NULL | 現在は`/admin`だけを保存する。CHECK制約で固定する。 |
| `created_at` / `expires_at` | TIMESTAMPTZ, NOT NULL | 発行時刻と10分後の期限。 |
| `consumed_at` / `invalidated_at` | TIMESTAMPTZ, NULL | 一回使用確保時刻／置換無効化時刻。 |

`reference_hash`と`state_hash`のunique index、保守用の`expires_at` indexを置く。`consumed_at`または`invalidated_at`が非NULLの行は認証に使えない。

## `admin_sessions`

| column | 型・NULL | 制約・用途 |
| --- | --- | --- |
| `id` | UUID, NOT NULL | primary key。Server内部識別子。 |
| `reference_hash` | BYTEA, NOT NULL | session cookie参照値のSHA-256。UNIQUE。 |
| `authorized_email` | TEXT, NOT NULL | OIDC検証済みメール。Server限定。 |
| `csrf_token_hash` | BYTEA, NOT NULL | 同期CSRF tokenのSHA-256。 |
| `created_at` / `last_mutation_at` | TIMESTAMPTZ, NOT NULL | 発行時刻／認可・CSRF成功状態変更時刻。 |
| `absolute_expires_at` / `idle_expires_at` | TIMESTAMPTZ, NOT NULL | 発行から8時間／最終状態変更から30分。 |
| `revoked_at` | TIMESTAMPTZ, NULL | logout等の失効時刻。 |

`reference_hash`のunique indexを認証lookupに、`absolute_expires_at`と`revoked_at`のindexを保守選択に使う。`idle_expires_at <= absolute_expires_at`をCHECK制約とする。`revoked_at`が非NULL、またはいずれかの期限が現在時刻以前の行は無効である。

## DB access map

| operation | transaction / access |
| --- | --- |
| `oauth_start` | 旧未使用行を無効化し、新規transactionをINSERTする一つのtransaction。 |
| `oauth_callback` | 条件付き`consumed_at`更新で一回使用を確保する短いtransaction。Google交換はその外で行い、成功時だけsessionを別transactionでINSERTする。 |
| `admin_session_bootstrap` | session参照。無効確定時の失効更新だけ短いtransaction。 |
| 管理状態変更ガード | session参照と、成功時のidle期限更新を後続業務更新と同じtransactionに含める。 |
| `admin_logout` | 有効sessionだけ`revoked_at`を更新する短いtransaction。 |
