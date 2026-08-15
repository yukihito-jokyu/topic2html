# DEC-ARCH-003 — 実装基盤・境界構造・外部本人確認・共通検証手順

状態: **承認済み（2026-08-15、L3）**

> 2026-08-15に利用者がGinの採用、`backend/`と`frontend/`の分離、BackendのClean Architecture準拠を明示指示したため、このDecisionを改定した。同日、Backendを`handler`・`usecase`・`repository`・`domain`へ明示分割することも利用者が承認した。

## 課題

DEC-ARCH-001は信頼済みWebアプリとリレーショナル永続化を、DEC-ARCH-002は生成HTMLの別origin隔離を承認している。一方で、FEAT-001以降を実装可能なTaskへ分解するには、Web実行基盤、DB接続方式、Google本人確認の外部依存境界、配置責務、依存固定と検証手順が必要である。

現行実装には初期のGo Serverがあるが、リポジトリ分割およびレイヤー構造への移行は未完了である。親リポジトリの`README.md`と削除前コミット`HEAD^`には、Go 1.26.5、Node.js 24.18.0/npm 11、React/Vite、PostgreSQL、`pgx`、Taskによる検証構成が記録されている。これは再利用可能な根拠ではあるが、削除済みの過去実装を自動復元・自動採用する根拠にはならない。

## 判断基準

1. 管理認証、秘密値、Codex app-server連携をブラウザおよび生成HTMLへ出さない。
2. 版、公開状態、生成試行、タグ、配置の関係と更新を整合的に保存できる。
3. 管理・公開・隔離表示というDEC-ARCH-001/002の責務境界を崩さない。
4. 依存取得、静的検査、単体・結合・ブラウザ検証を開発者とCIで再現できる。
5. 初期リリースに不要なサービス分割、ORM、全言語共通ビルド基盤を増やさない。
6. HTTPフレームワーク、DB、外部IdPなどの詳細が業務規則・ユースケースへ逆流しない。

## 選択肢

### A. GinをHTTP境界に限定した二Toolchain基盤を採用する（推奨）

- 管理画面はNode.js 24.18.0/npm 11、TypeScript、React、Viteでビルドする。Browserに渡るのは画面コードと明示的に公開可能な設定だけとする。
- リポジトリは`backend/`と`frontend/`をトップレベルで分割する。`backend/`は信頼済みWebアプリだけを、`frontend/`は管理UIのNode.js/TypeScript資産だけを含む。生成HTMLの隔離originは別の実行・配信境界であり、Frontendの静的資産と同一視しない。
- 信頼済みWebアプリはGo 1.26.5とGin v1.12.0で実行する。GinはHTTPのルーティング、middleware、request/response変換の外側境界に限定する。認証、認可、管理API、公開対象の解決、Codex app-server連携、DB接続をServer側へ置く。
- Backendは次の4層を明示したClean Architectureを採用する。`cmd/`はcomposition rootとして、設定、repository実装、usecase、handlerを接続して起動するだけとする。

| 層 | 所有する責務 | 許可する内向き依存 | 禁止する責務・依存 |
| --- | --- | --- | --- |
| `handler` | Gin routing/middleware、HTTP入出力変換、認証済み主体の受け渡し | `usecase`、`domain`の公開モデル | SQL、外部SDK、設定・秘密値、repositoryの直接利用 |
| `usecase` | アプリケーション操作、業務フロー、外部I/O port | `domain` | Gin、HTTP、PostgreSQL、Google、Codex app-server、環境変数、repository実装 |
| `repository` | usecase portの実装、PostgreSQL、Google、Codex app-serverとの外部I/O境界 | `usecase`のport、`domain` | HTTP request/response、Gin、業務フローの制御、handlerへの依存、環境変数の直接読取り |
| `domain` | 業務概念、不変条件、業務規則 | なし | handler、usecase、repository、cmd、フレームワーク、HTTP、DB、外部provider、環境変数 |

許可される依存は`handler → usecase → domain`、`repository → usecase/domain`、`cmd → handler/usecase/repository/domain`だけとする。外部I/Oはusecaseが定義するport経由でrepositoryへ反転させる。Server限定設定は`cmd/`が読み取り・検証し、repositoryやhandlerへconstructor注入する。HTTP request/response、SQL行、OAuth provider型、設定値はdomain/usecaseの公開モデルにしない。具体的なpackage内の型・ファイル・RouteはFeature DesignおよびImplementationで決める。FrontendはBackend内部、PostgreSQL、秘密情報に依存せず、同一originで公開されたHTTP契約だけに依存する。
- Node依存は`package.json`と`package-lock.json`、Go依存は`go.mod`と`go.sum`で固定する。Ginは`github.com/gin-gonic/gin v1.12.0`を直接依存として固定し、更新は互換性・脆弱性評価とlockfile検証を伴う変更として扱う。CIの取得は`npm ci`、`go mod download`、`go mod verify`を用いる。
- 永続化の正本はPostgreSQL 18系、Go接続は`github.com/jackc/pgx/v5`と`pgxpool`とする。ORMは採用せず、Parameter化SQLと明示Transactionを使う。スキーマ変更は版管理されたSQL Migrationとし、Migration runnerだけが適用する。
- Google本人確認は、Server側のOAuth 2.0 Authorization Code FlowとOpenID ConnectのID Token検証を使う。開始要求ごとに`state`（必要に応じて`nonce`とPKCE verifier）を生成・安全に保持し、callbackで`state`を照合してから処理する。ID Tokenは署名/JWKS、`issuer`、`audience`（複数の`audience`では`azp`も）、有効期限、`nonce`、`email_verified=true`を検証し、検証済みメールアドレスだけをServer限定の許可メールアドレス1件と完全一致で照合する。GoogleのToken、client secret、session署名鍵、DB接続文字列はBrowser・生成HTML・ログ・fixtureへ出さない。
- 配置責務は以下とする。具体的なHosting事業者、DNS、TLS終端、origin名、Network Policyは後続Decisionへ残す。

| 配置面 | 責務 | 禁止事項 |
| --- | --- | --- |
| 信頼済みアプリorigin | 管理画面の配信、Go Server、Google/Codex連携、管理・公開対象の解決 | 生成HTMLを同一DOM・同一originで実行しない。 |
| 生成HTMLの隔離origin | 解決済みHTMLだけをプレビュー・公開表示する | Cookie、管理API client、Google/Codex/DB資格情報、管理データを持たない。 |
| PostgreSQL | 管理データのSystem of Record | Browser・生成HTMLから直接到達させない。 |

- 共通の入口としてTaskを整備し、少なくとも次をCIでも同じ順序で実行する。
  - Frontend: `npm ci`、型検査、Biomeによる整形検査・静的検査、production build、Vitest。
  - Backend: Go Module検証、`gofmt`検査、`golangci-lint`、`go test ./...`、build。
  - 境界: 実PostgreSQLを使うMigration/Repository結合試験、PlaywrightによるGoogle認可のtest double経由の認可・公開・隔離E2E。通常CIへ実Google/Codex資格情報は渡さない。
- 構造: Backendの`domain`/`usecase`がGin、PostgreSQL、Google、Codex app-server、環境設定へ直接依存していないこと、`handler`がrepositoryを直接利用していないこと、repositoryがhandlerへ依存していないことを、依存検査または同等の静的検査で確認する。Gin経由のrequest/response変換はhandlerのHTTPテストで確認する。

**利点:** 利用者が指定したGinをHTTP境界へ限定し、秘密をGo Server側へ閉じたまま、Backendの業務規則を外部技術から分離できる。Frontend/Backendの責務が明確になり、関係データとBrowser境界を検証できる。

**注意点:** 既存の初期Go Serverは新しい配置・依存方向へ移行する必要がある。各Toolchain・DB・Google OAuth client・Ginの更新を運用する。

### A'（改定前）. 標準`net/http`を継続する

既存の初期Go Serverに近い構成として、標準`net/http`をHTTP境界に使い続ける。

**不採用とする理由:** 利用者がGinの採用を明示した。GinをHTTP adapterの外側に限定する本Decisionは、利用者指定を満たしつつ、Domain/Applicationをフレームワーク依存にしない。

### B. 単一のNode.jsフルスタック基盤に統一する

管理画面、Server、Google連携、DB接続をTypeScript/Node.jsで統一する。

**利点:** 言語と依存管理を一つにできる。

**不採用とする理由:** 親リポジトリに記録されたGo Server基盤と検証資産を捨てることになり、外部連携・永続化・隔離境界を新たに選び直す範囲が大きい。現在の初期リリース要件に、その移行コストを正当化する根拠はない。

### C. SQLiteを含む単一プロセスの最小構成にする

GoまたはNodeの1プロセスとSQLiteで開始し、DB ServerやMigrationの運用を後回しにする。

**利点:** ローカル開始時の構成が小さい。

**不採用とする理由:** 版、公開状態、試行、タグ、配置の関係を長期の正本として扱い、公開アプリとして運用する前提では、DB種別の後日移行が高コストになる。接続・Migration・バックアップの運用境界も曖昧になる。

## 推奨する決定

**選択肢Aを採用する。**

この決定は、実行時の責務、リポジトリ境界、Backendの依存方向、依存固定・検証の共通基盤だけを決める。個別Route、テーブル/列、Googleログイン画面、session cookieやCSRFの具体方式、Codex app-serverの接続プロトコル、隔離originの具体Host/Header/sandbox、Hosting事業者、Backendの個別package/file構成は後続のFeature Designまたは別Decisionで決める。

## 承認後の影響

- FEAT-001はGoogle OAuth/OIDCのServer側連携と、検証済みメールアドレスの完全一致認可を前提に詳細なTaskへ分割できる。Gin HTTP変換はhandler、OAuthのアプリケーション操作とportはusecase、Googleおよび設定境界はrepository、認可に関する不変条件はdomainに置く。
- FEAT-002〜005はPostgreSQL、Go/Gin Backend、React Frontend、隔離originという同じ責務境界を利用する。各Feature Designは、必要なdomain規則、usecase、外部portとrepository実装、handlerのHTTP変換を対応付け、禁止依存を確認する。
- 既存の初期Go Serverを`backend/`へ移行し、GinおよびClean Architectureの境界へ適合させる作業は、実装前レビューを経た専用の構造移行Taskとして扱う。既存の認証設定・Google境界の安全要件を後退させない。
- Task分解では、具体のfile path・関数名・framework APIを固定せず、承認済みの技術制約に沿う実装責務と受け入れ基準だけを記載する。
- 実Google/Codexへの接続確認は、秘密を隔離できる専用環境のsmoke testに限定し、通常CIではtest doubleを使用する。

## 見直し条件

- Codex app-serverがGo Serverから安定して接続できないことが一次資料または試作で判明した場合。
- PostgreSQLの容量・性能・運用条件が初期リリースの前提と合わない場合。
- DEC-ARCH-002の別origin隔離を、この配置でBrowser E2Eにより検証できない場合。
- ユーザーがHosting、対応環境、組織標準の技術基盤を別途指定した場合。
- Ginの更新または導入が、必要なGo版本体・セキュリティ・性能・運用要件と両立しないことが一次資料または試作で判明した場合。この場合も変更には利用者承認を要する。
