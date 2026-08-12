# ADR-0007: Go・Vitest・PostgreSQL・Playwrightを試験層に分ける

- 状態: 採用
- 決定日: 2026-08-12
- 対象: Test基盤

## 課題

Compileや型検査だけでは、HTTP Runtime入力、Database Transaction、BrowserのCookie/Storage/Origin、公開状態の反映を保証できない。Go BackendとTypeScript Frontendをそれぞれ速く確認する試験と、実Database/Browser境界を確認する試験を分け、S1の禁止到達を実際に観測できる基盤が必要である。

## 評価基準

1. Go BackendのUnit/HTTP Contract Testを標準Toolchainで実行できること。
2. React/Viteと同じ変換設定でFrontend Unit/Component Testを実行できること。
3. PostgreSQL 18を使うRepository/Migration Integration Testを実行できること。
4. 実BrowserでPreview/Publicの隔離とNetwork Requestを観測できること。
5. Test Doubleと実Service試験の役割を区別し、要件にないBrowser保証を追加しないこと。

## 候補

- Go `testing` / `httptest` + Vitest + PostgreSQL Integration + Playwright: 採用する。言語内の高速試験と実境界試験を役割分担できる。
- Go向け第三者Test Frameworkを必須化: 採用しない。表形式Test、Subtest、HTTP Handler試験は標準 `testing` / `net/http/httptest` で満たせ、独自Assertion規則を増やす必要がない。
- Jest + Go testing + Playwright: 採用しない。FrontendでViteとは別のTypeScript/Module変換設定を維持する必要がある。
- VitestとDOM Emulatorだけ: 採用しない。`DOM Emulator` はNode.js内でBrowser APIの一部を模倣する仕組みであり、実際のCookie、Origin、Navigation、Network隔離を保証できない。
- 手動Browser確認だけ: 採用しない。`RA-05`〜`RA-14` の回帰を繰り返し検出できない。

## 決定と採用理由

- Go標準 `testing` をBackendのUnit/Contract Test、`net/http/httptest` を `http.Handler` のRequest/Response試験に使う。`httptest` は実Portを常に開かずに標準HTTP処理を呼び、Status、Header、JSONを確認するための標準支援である。
- Vitest 4.1系列を管理FrontendのUnit、React Component、Frontend側Contract Testに使う。Go Backendの試験Runnerには使わない。
- Canonical Fixture（両言語が読む正本JSON例）に対し、Goの構造体/ValidatorとFrontendの型/Runtime検査を別々に実行する。成功例だけでなく、欠落・不正 `Content-Type`、未定義 `+json`、未知Field、欠落Field、禁止値、Body上限超過、余分なJSON値、内部Field混入の拒否例を含める。各Route/APIの既存機能Ownerは、自身が決めた有限Byte値をHandlerへ適用し、妥当な `application/json` Parameterの受理、成功・失敗応答の `Content-Type: application/json`、丁度1個のJSONだけの受理と併せてHTTP Testで確認する。L7-M2-S1のComposition OwnerはGlobal RouterへのMount時に、全JSON Routeへ有限capが存在することを検査する。ここでは共通Byte値、共通HTTP実装、機能固有Field、Status、Error値を定義しない。
- PostgreSQL 18の一時Test InstanceをGo Repository、Transaction、Migration Integration Testに使う。`Integration Test` は複数部品を実際につないだ境界確認であり、pgxのMockだけでは実SQL、制約、Lock、Rollbackを証明できない。Migrationは2つのRunnerを同じDatabaseへ競合させ、先行Runnerだけが適用すること、後行RunnerではDatabase側 `lock_timeout=10s` がcore `LockTimeout=15s` より先に発火してDriver `Lock()` が終了すること、DB lock timeout Error後に後続Migrationを実行しないことを時刻付きで確認する。さらに後行Runner終了後に遅延Lock取得がないこと、`pg_stat_activity` と `pg_locks` に専用Session/Advisory Lockが残らないこと、先行解除後の新しいRunnerが成功することを確認する。別Caseでcore `ErrLockTimeout` はDriver取消機能ではないことも固定し、これを正常な解放経路として扱わない。
- Playwright 1.62系列をBrowser E2Eに使い、Chromium、Firefox、WebKitのうち境界機能に関係するProjectをCIへ固定する。複数Engineの試験は隔離機能の実装差を検出するためであり、要件にない全Browser Supportを約束するものではない。
- Codex app-serverは通常CIでProtocol Test Double（同じ入出力契約を再現する代用品）を使い、実Service接続は秘密情報を管理できるL7 Smoke Testへ分ける。`Smoke Test` は主要経路が起動・接続できるかを絞って確認する試験である。

## 制約

- 許可: Deterministic（同じ入力なら同じ結果になる）なTest Double、一時PostgreSQL、PlaywrightのNetwork/Storage観測、失敗時Artifact保存。
- 禁止: DOM EmulatorやMockだけで隔離/永続化合格とすること、Production資格情報を通常CIへ入れること、外部Service通信を無条件に実行する不安定なTest。
- 条件付き: 実Codex/Google/外部Service試験は、隔離された資格情報、実行環境、費用・Rate Limit（一定時間に許される呼出回数の上限）管理がある場合だけ実行する。
- PlaywrightのBrowser Projectは試験範囲であり、製品の対応Browser要件ではない。
- HTML形式検証のTest Caseと具体方式はL3、公開/Preview隔離のE2E詳細はL6/L7が所有する。

## 必須試験責務

Reference/CI実行トポロジーは再現可能な参照環境と継続的試験環境の接続構成、TLSはTransport Layer Security（通信相手と通信内容を保護する暗号化規則）を意味する。実際のHost・Port・Origin・TLS・Proxy・Network Policy・Cacheを使う理由は、模擬した関数だけではCookie、通信経路、資格情報非付与を証明できないためである。本番Hosting製品や本番証明書の選定を意味しない。

| 層 | 主な確認 | 確認しないこと |
| --- | --- | --- |
| Compile / Type Check | Go Backend内、Frontend内それぞれの型整合 | 言語間Runtime契約、外部入力の安全性 |
| Go testing / httptest | 各Route/API OwnerがBackend状態遷移、DTO変換、Content-Type欠落/不正、Route固有Body上限超過、丁度1 JSON値、未知Field、成功/失敗JSON Response Header、安全なError応答、HTTP Method/Routeを確認。L7 Composition OwnerがMount済み全JSON Routeの有限capを確認 | 実PostgreSQL、実Browserの隔離 |
| Vitest | Frontend状態遷移、Canonical Fixtureとの契約、UI Component | Go Backend内部、実Browserの隔離 |
| PostgreSQL Integration | pgx Query、Migration順序/失敗停止、2 Runner競合、DB側10秒がcore 15秒より先に発火する時系列、後続非実行、遅延Lock取得なし、Session/Lock残留なし、解除後の新Runner成功、Transaction、版と公開参照の分離 | Browser到達性 |
| Playwright E2E | Preview/Public共通隔離、現在公開版だけの匿名表示、公開切替・非公開化、Cookie/Header/Storage/管理APIへの到達不能、外部通信にアプリ資格情報がないこと | 特定端末での視覚品質保証 |
| L7 Smoke | Reference/CI実行トポロジーの実Host・Port・Origin・TLS・Proxy・Network Policy・Cacheと実Codex等の主要接続 | 本番Hosting製品の選定、または全異常系の網羅 |

## Version固定方針

- Backend TestはGo 1.26.5標準 `testing` / `httptest`、Frontend TestはVitest 4.1系列、Browser TestはPlaywright 1.62系列を開始Baselineとする。
- Go版はBackend Toolchainと同じCI設定、Vitest/Playwrightの正確なPackage版は `package-lock.json` へ固定する。
- PlaywrightのBrowser Binaryは対応Packageと組になるため、CI Cacheだけを独立更新せず `playwright install` の管理対象版を使う。
- PostgreSQL Test Instanceとgolang-migrateはProductionと同じ18.4、4.19.1へ固定する。更新は全試験層の成功後に行う。

## 見直し条件

- Go標準Testだけでは必要な並行性・Contract検査を保守できないことが計測で判明する。
- VitestがVite/TypeScriptの採用系列をSupportしない。
- PlaywrightがL6で選ぶ隔離機構または必要Browser Engineを観測できない。
- 実BrowserとReference/CI環境の差により `RA-05`〜`RA-14` の重要保証を再現できない。
- Migration/Transaction TestがProduction同等PostgreSQLで実行できない。

## 公式一次資料

すべて2026-08-12確認: [testing package](https://pkg.go.dev/testing@go1.26.5)、[httptest package](https://pkg.go.dev/net/http/httptest@go1.26.5)、[MaxBytesReader](https://pkg.go.dev/net/http@go1.26.5#MaxBytesReader)、[mime.ParseMediaType](https://pkg.go.dev/mime@go1.26.5#ParseMediaType)、[golang-migrate v4.19.1 core](https://github.com/golang-migrate/migrate/blob/v4.19.1/migrate.go)、[v4.19.1 PostgreSQL Driver](https://github.com/golang-migrate/migrate/blob/v4.19.1/database/postgres/postgres.go)、[v4.19.1 Query Parameter filter](https://github.com/golang-migrate/migrate/blob/v4.19.1/util.go)、[lib/pq v1.10.9接続文字列](https://github.com/lib/pq/blob/v1.10.9/doc.go)、[PostgreSQL 18 lock_timeout](https://www.postgresql.org/docs/18/runtime-config-client.html#GUC-LOCK-TIMEOUT)、[PostgreSQL 18 Advisory Lock関数](https://www.postgresql.org/docs/18/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS)、[PostgreSQL 18 pg_locks](https://www.postgresql.org/docs/18/view-pg-locks.html)、[PostgreSQL 18 pg_stat_activity](https://www.postgresql.org/docs/18/monitoring-stats.html#MONITORING-PG-STAT-ACTIVITY-VIEW)、[Vitest 4 guide](https://v4.vitest.dev/guide/)、[Playwright release notes](https://playwright.dev/docs/release-notes)、[Playwright browser management](https://playwright.dev/docs/browsers)、[PostgreSQL 18 documentation](https://www.postgresql.org/docs/18/index.html)。
