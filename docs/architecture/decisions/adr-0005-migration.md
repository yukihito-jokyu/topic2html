# ADR-0005: Database MigrationをVersion付きSQLと単一Runnerで管理する

- 状態: 採用
- 決定日: 2026-08-12
- 対象: Migration

## 課題

Migrationは、保存済みデータを残したままDatabase Schema（表・列・制約などの保存構造）を新しい形へ移す仕組みである。開発者ごとの手作業やApplication起動時の暗黙変更では、適用順、失敗位置、実行済み版を追跡できない。Go Backendと独立して、保存構造を安全に移行し、複数Processが同時適用しない仕組みが必要である。

## 評価基準

1. 変更をVersion管理された順序付きSQL Fileとして残せること。
2. 適用履歴、未適用変更、失敗途中の状態を機械判定できること。
3. PostgreSQL Advisory Lock、Lock待ちの有限Timeoutと、可能な変更でのTransactionを扱えること。Advisory Lockは同じDatabaseで複数Runnerが競合したとき、一方だけにMigration実行を許す排他印である。Timeoutは待続けることとMigrationに失敗したことを区別するために必要である。
4. Go Modulesで固定でき、Go Backendと同じRepository/CIから明示Commandとして実行できること。
5. Down Migrationを常に安全と誤認させないこと。

## 候補

- golang-migrate/migrate v4: 採用する。Version付きの `up` / `down` SQL、適用Versionと `dirty` 状態、PostgreSQL Driver、CLI（Command Line Interface、Terminalから明示Commandで実行する窓口）とGo Libraryを提供する安定系列である。
- goose v3: 採用しない。SQL/Go Migrationを扱える有力候補だが、本製品はMigrationにApplication Codeを混ぜずSQLを正本にしたく、migrate v4で必要な順序・履歴・停止を満たせる。
- ORMの自動Schema同期: 採用しない。起動したApplication Instance（動作中のアプリ実体）が暗黙に変更し、順序とOwnerが不明瞭になる。
- 手書きSQLを運用者が都度実行: 採用しない。SQL自体は採用するが、履歴、単一実行、失敗位置、再現手順が共通化されない。
- Application独自Runner: 採用しない。既存Toolで満たせる履歴・順序・失敗停止を独自実装する必要がない。

## 決定と採用理由

`github.com/golang-migrate/migrate/v4` の4.19.1を採用し、Version付きの `.up.sql` / `.down.sql` を単一Migration Runnerから実行する。PostgreSQL Driverの `Lock()` はMigration適用前にSession-level（そのDatabase接続が続く間保持される範囲）の `pg_advisory_lock` を自動取得する。このLockはDriverの `Unlock()` またはDatabase Session終了まで保持され、別Processが誤って同時起動しても同じMigrationを並行適用させない。Application側がLockを「有効化」する別Switchがあるとはみなさず、Driverの実際のLock/Unlock動作を契約にする。Runnerは保存構造を変更する唯一の明示入口であり、通常のGo Backend起動処理はMigrationを自動適用しない。BackendとRunnerを分けるのは、複数Application Processの起動競合と、通常稼働資格情報へSchema変更権限を常時持たせることを避けるためである。

Lock待ちには、Migration専用PostgreSQL SessionのDatabase側 `lock_timeout=10s` と、coreの `Migrate.LockTimeout=15s` をこの順に作動させる開始Baselineを採用する。`lock_timeout` はDatabase ServerがLock取得を待つStatementを中止する設定である。golang-migrate v4.19.1が固定するlib/pq v1.10.9は接続文字列でPostgreSQL Runtime Parameterを設定でき、PostgreSQL Driverは `x-` で始まらないQuery Parameterをlib/pqへ渡す。そのためRunnerは秘密値を表示せず、Migration専用Database URLへ `lock_timeout=10s` を加える。Driverの初期化もMigration履歴Table確認のためLockを取るので、RunnerはDriverを開く前にURLの有効値を検査し、同じ接続文字列で短命なPreflight Session（本実行前の設定確認専用接続）を開いて `SHOW lock_timeout` が10秒であることを確かめ、閉じてからDriverを開く。値が欠落・不正ならMigrationを開始せず終了する。この順序が必要なのは、設定漏れのままDriver初期化時から無限待ちへ退行させないためである。

coreの15秒はDriverがLockを取得するまで呼出側が待つ上限に過ぎず、Driver処理の取消機能ではない。v4.19.1 coreは別Goroutine（Goで並行処理する実行単位）でDriver `Lock()` を呼び、15秒後に `ErrLockTimeout` を返してもそのGoroutineを取消さない。さらにPostgreSQL Driverは取消不能な `context.Background()` で `pg_advisory_lock` を実行する。このためcore 15秒だけを安全保証には使わない。Database側10秒が先に `pg_advisory_lock` StatementをErrorで終了させ、その結果Driver `Lock()` のGoroutineが終了してからcore 15秒へ到達する順序を必須とする。期待する競合時結果はDB lock timeout Errorであり、後続Migrationは1件も実行しない。Server上のStatementは既に中止されているので遅れてAdvisory Lockを取得せず、Runnerが専用Handle/Sessionを閉じることでSessionも残さない。もし `ErrLockTimeout` が返った場合はこの順序が壊れた異常であり、やはり後続を実行せず、専用Sessionを再利用せずRunnerを失敗終了する。Runnerは `Close()` だけで待機中Driverを取消せるとはみなさず、Database側10秒の発火とDriver処理終了を試験で確認してから終了する。

PostgreSQLの `statement_timeout` は個々のSQL Statementの総実行時間、Database側 `lock_timeout` はLockを待っている時間だけを制限する。golang-migrateの `x-statement-timeout` もMigration SQLを実行する `runStatement` のGo Contextへだけ適用され、Driver `Lock()` の `pg_advisory_lock` には適用されない。したがってこれらをAdvisory Lock待ちの代用にせず、Migration Sessionでは `lock_timeout=10s` を必須にし、`statement_timeout` を設定する場合はLock経路で10秒より先に発火しない値とする。Issue #7は10秒・15秒の大小関係を保った環境別設定と秘密非露出を所有する。

一度共有環境へ適用したMigration Fileは書き換えず、新しいVersion Fileで前進修正する。PostgreSQLがTransaction内実行を許す操作は、対応する `.up.sql` 内に `BEGIN` / `COMMIT` を明示し、File全体を一つの成功・失敗単位にする。golang-migrateが全SQLを自動的にTransaction化するとはみなさず、SQL側で境界を目に見える形にすることが必要である。Transaction外操作が必要な場合は、理由、失敗時復旧、再実行条件を当該Migrationと後続設計へ記録する。

`dirty` はMigrationが完了せず、記録されたVersionの途中で停止した状態を表す。Runnerはdirty状態で後続Migrationを続けず、人が原因とDatabase状態を確認する。`force` は履歴Versionを強制変更する操作であり、通常の自動復旧には使わず、L1-M2-S3が実Database状態、復旧手順、承認を記録した場合だけ許可する。

`down` は前のSchemaへ戻す処理だが、データ削除を伴う変更は元情報を復元できないため、常に安全とは限らない。不可逆変更は安易なdownを用意せず、Backup/Restoreまたは前進修正手順をMigration統治OwnerであるL1-M2-S3が決める。論理Schemaの移行方針と不可逆変更の復旧条件を同じOwnerへ集め、機能ごとに安全基準が分かれることを防ぐためである。

## 制約

- 許可: Version管理されたSQL Migration、可能な変更でのTransaction、Driverが自動取得するPostgreSQL Advisory Lockを使う単一Runner、Database側10秒・core側15秒の有限二段Timeout、前進修正。
- 禁止: Application起動時の暗黙自動変更、共有環境適用済みFileの改変、複数Runnerの並行適用、DB lock timeout Errorまたは `ErrLockTimeout` 後のMigration続行、接続未解放の終了、dirty状態の自動続行、未記録の `force`。
- 条件付き: Transaction外操作、不可逆変更、`force` は、理由・復旧・再実行方法と承認を記録した場合だけ許可する。
- Migration Database URLはRunnerだけが実行時に読み、Command引数、Log、CI Artifactへ値を出さない。具体的な秘密設定方式、Database側 `lock_timeout`、core `LockTimeout`、`statement_timeout` 等の環境別Migration実行設定はIssue #7が決める。10秒 < 15秒の開始Baselineを変更する場合も、両方が有限でDatabase側が先に発火する大小関係を保ち、競合Integration Testで安全停止を再確認する。
- Table/Column/Index、命名規則の細部、Downtime、Backup、Downgrade方針はMigration統治OwnerのL1-M2-S3、専用接続の設定検証、Lock/Timeout順序、後続非実行、解放を含むRunner実装とIntegration TestはL1-M4-S2の責任とする。Issue #7はその方針や実装を独自に変えず、環境別設定値と秘密非露出を所有する。

## Version固定方針

- golang-migrate/migrate 4.19.1を開始Baselineとし、Go Library利用時は `go.mod` / `go.sum`、CLI Binary利用時は配布版と公式Checksumへ正確な版を固定する。
- v5へ自動更新せず、履歴Table、dirty状態、PostgreSQL Driver、既存SQL Migration、CLI終了Codeの互換試験を行う。
- Go 1.26.5、PostgreSQL 18.4、migrate 4.19.1、lib/pq 1.10.9、Database側 `lock_timeout=10s`、core `LockTimeout=15s` の組合せをCI Integration Testで固定する。

## 見直し条件

- v4が採用GoまたはPostgreSQL 18をSupportしない。
- Securityまたは保守上、v5以降への移行が必須になる。
- 複数Databaseや複数Schemaへ分割され、単一履歴・単一Runnerでは安全に適用できない。
- Transaction外の長時間変更が常態化し、Online Migration専用機構が必要になる。
- L1-M2-S3のMigration統治が別の保証を要求する。

## 公式一次資料

すべて2026-08-12確認: [golang-migrate v4.19.1 release](https://github.com/golang-migrate/migrate/releases/tag/v4.19.1)、[v4.19.1 core migrate.go](https://github.com/golang-migrate/migrate/blob/v4.19.1/migrate.go)、[v4.19.1 PostgreSQL Driver source](https://github.com/golang-migrate/migrate/blob/v4.19.1/database/postgres/postgres.go)、[v4.19.1 Query Parameter filter](https://github.com/golang-migrate/migrate/blob/v4.19.1/util.go)、[v4.19.1 go.mod（lib/pq 1.10.9固定）](https://github.com/golang-migrate/migrate/blob/v4.19.1/go.mod)、[lib/pq v1.10.9接続文字列](https://github.com/lib/pq/blob/v1.10.9/doc.go)、[v4.19.1 CLI documentation](https://github.com/golang-migrate/migrate/tree/v4.19.1/cmd/migrate)、[PostgreSQL 18 lock_timeout / statement_timeout](https://www.postgresql.org/docs/18/runtime-config-client.html#GUC-LOCK-TIMEOUT)、[PostgreSQL 18 Advisory Lock関数](https://www.postgresql.org/docs/18/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS)。
