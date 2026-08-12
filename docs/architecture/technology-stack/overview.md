# 横断技術構成

- 対象作業: `L1-M1-S2`
- 状態: 採用決定
- 確認日: 2026-08-12

## 1. この文書の目的

この文書は、topic2htmlを実装するときに全機能が共通して使う技術と、その技術だけでは保証できない安全条件をまとめる。Architecture Decision Record（`ADR`、設計上の重要な選択について理由・代案・制約・見直し条件を残す決定記録）の詳細は `docs/architecture/decisions/` に置く。

ここでいう「横断」は、特定の画面や一つの機能だけでなく、管理画面、Server、保存、Build、Testに共通するという意味である。この作業で実装コードや依存Packageは追加しない。実際のDirectory、Module、依存方向は後続Issue #6、設定値と秘密情報の扱いはIssue #7が具体化する。

### この文書で使う共通用語

- Frontendは利用者のBrowserで表示・操作を受け持つ部分、BackendはServerで認可・業務処理・保存接続を受け持つ部分である。FrontendはTypeScript、BackendはGoを使う別Toolchainであり、両者を分けるのはBrowserへ渡した情報が利用者や生成HTMLから観測でき、秘密を保持できないためである。
- ClientはAPIを呼ぶ側、Serverは要求を受けて応答する側である。APIはApplication Programming Interface（Program同士の操作窓口）、RouteはHTTP MethodとURLで定まる入口、Handlerは一つのRoute要求を受けて応答を作る処理、DTOは境界で許可したFieldだけを運ぶデータ、SchemaはField・型・必須性・許容値を定める規則である。
- HTTPはWeb上でRequest（要求）とResponse（応答）を交換する通信規則、Headerは本文と別に付ける制御情報、JSONは名前付きFieldと値を表すデータ形式である。Media TypeはBodyの形式を表す `Content-Type` の主値、Parameterはその後ろに付く追加条件である。JSON Routeは `application/json` と省略またはUTF-8の妥当なParameterだけを受理し、別契約を表す未定義 `+json` は受理しない。
- CompileはSource Codeを型検査して実行可能な成果物へ変換する工程である。Go構造体はFieldと型をCompile時に定めるBackendのデータ型、ValidatorはDecode後に必須性・長さ・列挙値・Field間条件を検査する処理である。構造体だけでは製品上の条件をすべて表せないため、双方を境界で使う。
- Canonical FixtureはFrontendとBackendが同じ契約か検査する正本JSON例である。TypeScript型をGoへ直接共有できないため、成功例と拒否例を両言語のContract Testへ渡す必要がある。
- Pointerは別の正本を指す参照値、Portは機能が必要とする操作を抽象化した窓口、AdapterはPortをDatabaseや外部Serviceへ接続する実装である。機能の意味と外部技術を同一Moduleへ閉じ込めないために分ける。
- Capabilityは特定操作だけを許す権限情報である。保持者へ管理者全権を渡さず、用途と有効範囲を限定するために使う。具体形式はL6で決める。
- Previewは管理者が未公開版を確認する表示、Publicは匿名利用者が現在公開版を見る表示である。対象解決の権限は異なるが、生成HTMLの隔離強度は同じでなければならない。
- CookieはBrowserが対象Web Originへの要求へ保存値を付ける仕組み、Authorization HeaderはHTTP要求へ認証情報を付ける項目、Tokenは権限や本人確認を示す秘密値、Client StorageはBrowser側の保存領域である。生成HTMLへ渡すと管理権限へ到達し得るため、すべて禁止対象になる。
- Transactionは複数のDatabase更新を一つの成功・失敗単位にする仕組み、Connection PoolはDatabase接続を上限内で再利用する仕組み、Cacheは正本の結果を一時保存して再利用する仕組みである。Poolは資源枯渇を防ぎ、Cacheは古い公開状態を返さない失効規則を必要とする。Advisory LockはApplication間で共通の数値Keyを使うDatabase排他である。MigrationではDatabase側 `lock_timeout` がLock待ちStatementを中止し、golang-migrate coreの `LockTimeout` は呼出側の待機だけを打ち切る。この二つはDriver処理を取消すかどうかが異なり、二重適用と無限待ちを安全に防ぐにはDatabase側を先に発火させる必要がある。
- Browser SandboxはWeb内容に許す機能を制限する仕組み、Private Networkは公開Internet外のNetwork、Loopbackは実行中の機器自身を指すNetwork宛先である。外部通信を許可しても、これらを管理領域への迂回路にしてはならない。
- 表示枠はBrowserが一つのHTML文書を表示・実行する文脈であり、配備Processを包む実体ではない。HostはHTTP要求を受ける配備先のMachineまたはServiceであり、Web Originに含まれるHost名だけを指す語ではない。ContainerはApplication Processと依存物を隔離して配備する実行単位であり、HTMLの表示枠を指さない。ProxyはBrowserと外部Serviceの間でHTTP要求を中継するServer側の役割であり、HTMLを表示するRendererとは異なる。Network Policyは配備環境で通信先・方向・Portを許可または禁止する規則であり、利用者の管理認可を代行しない。TLSはTransport Layer Security（通信相手と通信内容を保護する暗号化規則）であり、証明書を発行・運用するHosting業務そのものではない。この区別が必要なのは、L6が表示・通信の論理契約を決め、L7がReference/CI実行トポロジー（再現可能な参照環境と継続的試験環境の接続構成）の実Host・Port・Origin・TLS・Proxy・Network Policy・Cacheへ写して検証するためである。L7がContainer製品や本番Hosting製品を選ぶという意味ではない。
- Manifestは利用依存と基準版を宣言するFileであり、Frontendでは `package.json`、Backendでは `go.mod` を指す。Lockに相当する検証FileはFrontendの `package-lock.json` とBackendの `go.sum` で役割が異なる。前者は解決済みPackage版、後者はGo Module取得物のChecksumを記録する。CIはContinuous Integration（変更ごとにBuildとTestを自動実行する仕組み）である。

## 2. 採用構成

| 領域 | 採用技術 | 日本語での役割 | 必要な理由 |
| --- | --- | --- | --- |
| Backend言語/実行 | Go 1.26.5 | Server側HTTP、業務処理、Database/Codex接続をCompileする型付き言語 | 利用者の選択に従い、秘密をServer側へ閉じ、標準機能中心の小さい実行構成にするため |
| Backend HTTP | `net/http` / `http.ServeMux` | HTTP要求を受け、Method・Host・PathごとのHandlerへ振り分けるGo標準機能 | 現在のRoute要件を第三者Frameworkなしで明示するため |
| Backend入出力検査 | `mime.ParseMediaType` + `http.MaxBytesReader` + Go構造体 + `encoding/json` + 明示Validator | JSONのMedia Type、有限Body長、丁度1つの値、型、製品上の必須性・許容値を実行時に検査し、成功/失敗Responseを `application/json` と明示する | 形式誤認、過大入力、未知Field、禁止値、内部Field混入を共通境界で拒否するため |
| Frontend Toolchain | Node.js 24.18.0 LTS + TypeScript 6.0 | 管理画面Source、型検査、Browser成果物を作るFrontend専用環境 | Go Backend実行時と混同せず、React/Viteの公式Toolchainを固定するため |
| Frontend | React 19.2 | 管理画面の状態と画面部品を組み立てるUI Library | 管理操作を部品単位で扱うため。生成HTMLの描画には使わない |
| Frontend Build | Vite 8.2 | 開発時配信とBrowser向け成果物作成を行うBuild Tool | React/TypeScriptの経路を小さい構成で共通化するため |
| 言語間契約 | Field定義 + Canonical Fixture + Contract Test | GoとTypeScriptが同じJSONを許可/禁止するか照合する | Source型を直接共有できない二言語間のRuntime不一致を検出するため |
| 永続化 | PostgreSQL 18.4 | HTML版、試行、公開状態等の関係と更新をTransactionで保存する関係Database | 版と現在公開版を分離し、更新を一つの成功・失敗単位にするため |
| DB接続 | pgx 5.10.0 / `pgxpool` | GoからParameter化Query、Transaction、上限付き接続再利用を行うDriver | 自動生成SQLや暗黙Schema変更を避け、保存境界を明示するため |
| Migration | golang-migrate/migrate 4.19.1 + DB `lock_timeout=10s` + core `LockTimeout=15s` | Version付きSQLを順番に適用し、PostgreSQL DriverがSession-level Advisory Lockを自動取得するRunner | 二重適用を防ぎ、DBが先に待機Statementを中止してDriver処理を終了させ、後続非実行とSession解放を保証するため |
| 依存管理 | Go Modules + `go.sum`、npm 11 + `package-lock.json` | BackendとFrontendそれぞれの正規手段で依存版と取得物を固定する | 開発者とCIで二Toolchainを再現するため |
| Backend Unit/HTTP Test | Go `testing` / `httptest` | Go処理とHTTP Handlerを標準Toolchainで検査する | Backendの試験へFrontend用Node.js Runnerを持ち込まないため |
| Frontend Unit/Component Test | Vitest 4.1 | 管理画面処理、UI部品、Frontend側契約を検査する | ViteのTypeScript変換設定を共有するため |
| Browser E2E Test | Playwright 1.62 | 実Browserで入口から結果までを確認する試験基盤 | 生成HTML隔離、公開状態反映、外部通信をBrowser環境で確認するため |

各Versionは開始Baselineであり、任意の新しい版を取得してよいという意味ではない。正確な版をManifest/検証FileとCIへ記録し、更新ごとに該当する全試験を通す。

詳細な比較と決定理由は次の記録にある。

- [ADR-0001: Web実行](../decisions/adr-0001-web-execution.md)
- [ADR-0002: Frontend](../decisions/adr-0002-frontend.md)
- [ADR-0003: Backend](../decisions/adr-0003-backend.md)
- [ADR-0004: 永続化](../decisions/adr-0004-persistence.md)
- [ADR-0005: Migration](../decisions/adr-0005-migration.md)
- [ADR-0006: Build](../decisions/adr-0006-build.md)
- [ADR-0007: Test](../decisions/adr-0007-test.md)

## 3. Web実行の全体像

管理画面と生成HTMLは同じWeb製品の機能だが、同じ実行文脈には置かない。

1. Go Backendの表示対象解決処理が、管理者Previewなら指定された形式検証合格版、匿名Publicならその時点の現在公開版を解決する。
2. Backendは解決済みHTMLと表示に必要な非秘密情報だけを生成HTML Renderer（未信頼HTMLだけを表示する論理的な実行面）へ渡す。
3. RendererはReact管理画面のDocument Object Model（`DOM`、Browser内の画面要素Tree）へ生成HTMLを挿入せず、管理用Cookie、Token、API、Storage、Databaseへ到達できない実行文脈でHTMLを動かす。
4. 生成HTMLから外部HTTPS通信を一律には禁止しない。ただし、アプリは資格情報や管理データをその通信へ自動付与しない。

これは論理責務と禁止事項を固定する。具体的なWeb Origin、表示枠、Header、Browser Sandbox、Cache、Capability、Proxy使用有無はIssue #4の `DD-005` / `DD-007` に従いL6が決める。L7はL6の契約をReference/CIの実Host・Port・Origin・TLS・Proxy・Network Policy・Cacheへ反映し境界を検証する。Container製品と本番Hosting製品は本Taskでも現在のL7でも選定せず、本番DNS、証明書発行、CDN、監視もL7に含めない。配備要件が生じた場合だけ、Ownerを明示した後続ADRで選定する。

| 要件 | 本選定での扱い |
| --- | --- |
| `BR-021`・`NFR-003`・`AC-029` | 外部通信を一律禁止せず、アプリ資格情報・管理データの自動付与と管理領域への到達を禁止する |
| `NFR-002`・`AC-019` | PreviewとPublicを同じ制約のRendererへ渡し、管理文脈から隔離する |
| `CON-001` | 生成先はCodex app-serverへ固定し、別生成APIへの差替えを設計前提にしない |
| `NFR-004` | 選定理由へ利用者向け生成時間の上限・目標を追加しない |
| `NFR-005` | Playwright試験対象を、特定Browserや端末の表示保証と解釈しない |
| `NFR-006` | 技術採用を根拠に未定義のAccessibility合格指標を追加しない |

## 4. S1論理境界との正方向照合

「正方向照合」はS1で要求された各境界を採用構成がどこへ引き継ぐかの確認である。`TI` はTrust Invariant（信頼境界を越えても守る不変条件）の識別子である。L7の検証はReference/CIとBrowser/Network観測を対象とし、本番Hosting選定を意味しない。

| S1条件 | 本選定で固定する対応 | 具体保証のOwner |
| --- | --- | --- |
| `TI-01` 生成HTMLは未信頼 | React管理DOMへ挿入せず専用Rendererの入力とする | L6が隔離方式、L7がReference/CIで到達性を検証 |
| `TI-02` 最小データだけを渡す | Go BackendのPublic/Previewを別DTO・Validatorにし、解決済みHTMLと列挙した最小情報だけをEncodeする。全JSON RouteはMedia Type・有限長・丁度1値を検査する | 各Route/APIの既存機能OwnerがRoute固有Byte値とHandler適用を所有し、L7-M2-S1がMount済み全JSON Routeの有限capを検査する。L6はResolver/Renderer契約を所有する |
| `TI-03` 管理領域へ到達不可 | Rendererから管理Frontend/API、Database、資格情報、管理Storageへの経路を設けない | L6が具体経路、L7が到達不能を検証 |
| `TI-04` 外部通信は許容し資格情報を付けない | 外部HTTPSを一律拒否せず、管理Cookie・Authorization Header・秘密情報の自動付与を禁止する | L6が通信方式、L7が送信内容を検証 |
| `TI-05` 匿名は現在公開版だけ | Public Resolverを管理APIと分け、現在公開版以外を返さない | L4が不変版本文、L6が現在公開Pointerと解決 |
| `TI-06` Previewも同じ隔離 | 対象解決だけを管理経路で行い、実行はPublicと同じ制約のRendererへ渡す | L6 |
| `TI-07` 候補は自動で版等にならない | Codex出力を未信頼候補とし、技術構成は状態遷移を短絡しない | L3が生成・形式検証、L4/L5が保存・分類、L6が現在公開対象化を防ぐ |
| `TI-08` 失敗処理を共通化 | Go Handlerは共有された安全な失敗DTOだけを `Content-Type: application/json` で返し、内部Error/LogをEncodeしない | L1-M2-S2が機能間の安全な失敗契約、L3がCodex固有の再試行・失敗分類。契約Ownerと発生機能Ownerを分け、固有詳細を共有応答へ漏らさない |
| `TI-09` 新版で公開状態を変えない | PostgreSQLでも不変版記録と現在公開Pointerを別の所有・更新単位とする | L4が不変版、L6が現在公開Pointerと公開反映 |
| `TI-10` タグ・配置を版/公開切替と分離 | 関係Databaseを採用するが具体論理Schemaは決めない | L4が版、L5がタグ・配置 |
| `TI-11` 本人確認と管理認可を分離 | Backend内でも認証結果と管理許可判定を別責務にする | L2 |
| `TI-12` Codex app-server固定 | Go AdapterからCodex app-serverを使い、別生成APIへの差替えを決定しない | L3 |

## 5. 禁止到達先との逆方向照合

「逆方向照合」は採用技術が禁止経路を再導入していないか到達先から確認することを指す。`RA` はReachability Assertion（到達可否の表明）の識別子である。

| 禁止対象 | この段階で禁止する構成 | 残る保証 |
| --- | --- | --- |
| `RA-05` 資格情報・Session・認証情報 | Rendererへ管理Cookie、Authorization Header、Token、秘密設定を渡す | L6のOrigin/Header/通信設計とL7試験 |
| `RA-06` 管理DOM・状態・Client Storage | React `dangerouslySetInnerHTML`、管理DOMへの挿入、管理画面と同じStorage前提 | L6の実行面設計とL7 Browser試験 |
| `RA-07` 管理操作 | Rendererから管理Callbackや双方向命令Channelを設ける | L6のCapability設計 |
| `RA-08` 管理/内部API・認証Service | Public/Preview Rendererが管理API Clientを共有する | L2のAPI契約、L6の経路制御、L7試験 |
| `RA-09` DB・秘密情報 | BrowserへDB接続情報やBackend秘密情報をBuild時注入する | L2〜L6が用途別Port越しに永続化を使い、Issue #6がDatabase Module配置とFrontend/Rendererの依存禁止、Issue #7が資格情報非露出、L7が成果物とNetworkを検査 |
| `RA-10` 未公開版・履歴・試行・内部Log | Public Resolverが汎用Record/Errorを直接Encodeする | L4の用途別読取、L6のPublic Resolver |
| `RA-11` タグ・配置・公開状態の管理元データ | Rendererへ管理用編集Modelを渡す | L5の管理契約、L6の最小表示契約 |
| `RA-12` Codex・資格情報 | Codex Clientや接続情報をFrontend Bundleへ含める | L3のServer側Go Adapter、L7成果物検査 |
| `RA-13` Google資格情報 | Google TokenをRendererやPublic契約へ含める | L2、L7試験 |
| `RA-14` 現在公開版以外 | 匿名要求で版IDを指定し任意版を解決できる | L6の現在公開Pointer/Public Resolver、L4の不変版Content Port、L7試験 |

外部Serviceへの到達 `RA-03` は許可し、閲覧者自身のIP Address等を外部Serviceが観測し得る `RA-04` は要件上受容する。管理領域への戻り、資格情報付与、Private Network/Loopback到達を防ぐ具体方式は、Browser直結かServer ProxyかをL6で選んだ後に検証可能な規則へする。

## 6. 後続レベルへ残す保証

| Level | ここで渡す固定条件 | ここでは決めない事項 |
| --- | --- | --- |
| L2 | 本人確認と管理認可の分離、認証/Session秘密情報はGo Server側だけ | Google認証Flow、Cookie、CSRF対策、個別管理API Field |
| L3 | Go BackendからCodex app-serverを使い、出力を未信頼候補とする | 接続、Process、Prompt、再試行、形式検証の具体方式 |
| L4 | PostgreSQL上の不変版正本。現在公開PointerはL6へ分離 | Table、Column、Index、Repository、版保存の具体方式 |
| L5 | タグ・配置を版および公開切替から分離 | タグ・配置Schemaと操作API |
| L6 | 現在公開Pointer、Resolver/Renderer分離、Preview/Public共通隔離、公開反映、管理文脈への戻り遮断、外部通信と資格情報の分離 | Origin、Header、Sandbox、Cache、Capability、Proxy使用有無。L7はReference/CIへの反映と境界検証だけを担う |

横断するField定義、Canonical Fixture、Port、言語別DTO/Validatorの物理実装とMigration Runnerは、承認済みTask Mapにある各L1-M4 Taskが自身の対象範囲で所有し、Migration統治方針はL1-M2-S3が所有する。全JSON Routeに有限Body上限を必須とし、各Route/APIの既存機能OwnerがRoute固有Byte値、Handler適用、HTTP Testを所有する。L7-M2-S1のComposition OwnerはMount時に全JSON Routeの有限capを検査する。L1-M2-S2はHTTP・Serializationを対象外、L1-M4-S1はHTTP/RPC Handlerを対象外なので、共通上限値、Route別Byte値、共通HTTP実装を割り当てない。本ADRも横断共通の既定Byte数や共有実装Ownerを発明せず、共有化には承認済みTask MapのOwner変更またはTask追加を先に必要とする。Migration統治はL1-M2-S3、Database側10秒がcore 15秒より先に発火するRunner実装・解放・試験はL1-M4-S2、環境別設定と秘密非露出はIssue #7が所有する。L7はReference/CI、Browser、成果物、Network観測で検証する。Container製品や本番Hosting製品が必要になった場合はOwnerを確定した後続ADRで決める。

## 7. 後続Issueへの引継ぎ

Issue #6は次の構成要素をModule/Directoryへ配置し、依存方向と変更Ownerを一意にする。構成要素は実行責務のまとまりであり、まだFile名を意味しない。

| 構成要素 | 許可する依存 | 禁止する依存 |
| --- | --- | --- |
| Admin Frontend | 管理用TypeScript DTO/API Client、Rendererへの不透明な導線 | Database、Codex、秘密設定、生成HTML DOM |
| Go Backend HTTP | 各機能Ownerが所有する用途別Handler、Go DTO/Validator、Media Type・Route固有Body上限・丁度1 JSON・Response Content-Typeの検査を、Issue #6がModule/Directoryへ配置して依存方向を固定する。Issue #6は共通HTTP実装やByte値のOwnerにはならない | 上限のないDecode、未定義 `+json`、未信頼Schemaの動的実行、Database Recordの直接応答 |
| Contract基盤 | 言語非依存Field定義、Canonical Fixture、Go/TypeScriptのContract Test | 一方の言語型だけを正本とすること |
| Preview/Public Resolver | L4の不変版Content Port、L6の公開Pointer/Capability | Rendererから管理機能へ戻るPort |
| Credentialless Renderer | 解決済みHTML、列挙済み最小表示情報 | 管理API Client、認証状態、Database、Codex、管理Storage |
| PostgreSQL Adapter | FeatureごとのRepository Port、pgx/pgxpool | Browser/Rendererからの直接参照 |
| Migration Runner | golang-migrate、Version付きSQL、`lock_timeout=10s` を接続時に検証する専用PostgreSQL Session、Advisory Lock、core `LockTimeout=15s`、Driver終了後の接続解放 | 通常Backend起動による暗黙Schema変更、DB lock timeout Errorまたは `ErrLockTimeout` 後の後続Migration、core Timeoutだけに依存した取消 |
| Test基盤 | Go testing、Vitest、PostgreSQL Test Instance、Playwright、Protocol Test Double | Production秘密情報を使う通常CI |

`RA-09` は一つのAdapterだけでは閉じない。L2〜L6は所有データの意味を保って用途別Port越しにPostgreSQLを利用し、Issue #6はPort/AdapterのModule配置と禁止依存、Issue #7はDatabase資格情報をBackend/Migration Runner外へ露出しない設定境界、L7はBundle・Reference/CI通信・試験成果物を検査する。保存規則、物理依存、秘密情報、実行時検査のどれか一つでも欠けると到達不能を証明できないためである。

Issue #7は次の構成主体ごとに設定と秘密情報の露出可否を分類する。「主体」は設定を読み利用する実行上の当事者を指す。

| 構成主体 | 必要となる設定の分類 | 絶対に渡さないもの |
| --- | --- | --- |
| Admin Frontend | Browser公開可能と明示された非秘密設定だけ | Google Token、Session秘密、Database/Codex資格情報 |
| Go Backend | HTTP、Database、Google、CodexのServer側設定 | 応答やFrontend Buildを通じた秘密値の再公開 |
| PostgreSQL Adapter / Migration Runner | Database接続、Pool、Migration専用DB `lock_timeout=10s`、core `LockTimeout=15s`、PostgreSQL `statement_timeout` 等の相互に対象が異なる実行設定。Issue #7は値を環境へ与え秘密を隠すが、Runnerの実装Ownerではない | Browser Bundle、Renderer、Client Storage、Log/Artifactへの秘密値、無限Lock待ち、10秒 < 15秒の順序破壊 |
| Codex app-server Adapter | Server側の接続/Process設定 | Frontend/Renderer、匿名応答 |
| Preview/Public Resolver | Route、Capability、公開解決に必要なServer設定 | 管理SessionのRendererへの転送 |
| Credentialless Renderer | L6が列挙する非秘密の表示面設定だけ | Cookie、Authorization Header、秘密環境変数、管理元データ |
| Test基盤 | Test専用の隔離設定と代用品。実Service Smokeだけ条件付き資格情報 | Production資格情報を通常CI Log/Artifactへ出すこと |

## 8. 版固定と更新規則

- Backendは `go.mod` で最低必要Go版1.26.5を宣言し、CIと開発手順で実行ToolchainのGo 1.26.5を固定する。第三者Moduleは `go.mod` / `go.sum` でVersionと取得物を検証し、1.26系Patch更新もGo Test、Integration、Browser E2E後に取り込む。
- FrontendはNode.js 24.18.0/npm 11を `package.json` / CI、依存を `package-lock.json` へ固定し、`npm ci` で再現する。Node.jsはFrontend Toolchain専用でありBackend Runtimeではない。
- React 19.2、TypeScript 6.0、Vite 8.2、Vitest 4.1、Playwright 1.62、pgx 5.10.0、golang-migrate 4.19.1（内部PostgreSQL接続Driverはlib/pq 1.10.9）は採用系列内の更新でもBuild・Unit・Contract・Integration・Browser E2Eを通す。
- PostgreSQLは公式18.4 Releaseを根拠に18.4から開始し、18系の後続Patchへ試験後に追随する。Major 19へ自動更新しない。pgxは公式v5.10.0 ReleaseとVersion指定Package文書に固定し、可変の `master` だけを版根拠にしない。
- Go、Node.js、PostgreSQL、Browser Engine、Codex app-serverの実行版は診断情報へ記録可能にする。記録Fieldは後続が設計する。

## 9. 共通の見直し条件

- 採用系列が公式Support終了へ入る、または重大な脆弱性修正を受けられない。
- Codex app-serverの必須契約をGo Backendから扱えない。
- Go/TypeScript間のContract FixtureでRuntime整合を保てない。
- L6の具体設計で `RA-05`〜`RA-14` を満たせない。
- 外部通信許可と資格情報非付与、またはPreview/Public同一隔離を両立できない。
- Migrationの順序・単一実行・失敗時停止を一Runnerで保証できない。
- 試験基盤が実Browser/Database境界を検証できない。

## 10. 公式一次資料

すべて2026-08-12に確認した。

- [Go Release History](https://go.dev/doc/devel/release)
- [Go 1.26 release notes](https://go.dev/doc/go1.26)
- [Go Modules Reference](https://go.dev/ref/mod)
- [net/http package](https://pkg.go.dev/net/http@go1.26.5)
- [MaxBytesReader](https://pkg.go.dev/net/http@go1.26.5#MaxBytesReader)
- [mime.ParseMediaType](https://pkg.go.dev/mime@go1.26.5#ParseMediaType)
- [encoding/json package](https://pkg.go.dev/encoding/json@go1.26.5)
- [testing package](https://pkg.go.dev/testing@go1.26.5)
- [httptest package](https://pkg.go.dev/net/http/httptest@go1.26.5)
- [pgx v5.10.0 release](https://github.com/jackc/pgx/releases/tag/v5.10.0)
- [pgx v5.10.0 package](https://pkg.go.dev/github.com/jackc/pgx/v5@v5.10.0)
- [pgxpool v5.10.0 package](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool@v5.10.0)
- [golang-migrate v4.19.1](https://github.com/golang-migrate/migrate/releases/tag/v4.19.1)
- [golang-migrate v4.19.1 core](https://github.com/golang-migrate/migrate/blob/v4.19.1/migrate.go)
- [golang-migrate v4.19.1 PostgreSQL Driver](https://github.com/golang-migrate/migrate/blob/v4.19.1/database/postgres/postgres.go)
- [golang-migrate v4.19.1 Query Parameter filter](https://github.com/golang-migrate/migrate/blob/v4.19.1/util.go)
- [golang-migrate v4.19.1 go.mod](https://github.com/golang-migrate/migrate/blob/v4.19.1/go.mod)
- [lib/pq v1.10.9接続文字列](https://github.com/lib/pq/blob/v1.10.9/doc.go)
- [PostgreSQL 18 lock_timeout / statement_timeout](https://www.postgresql.org/docs/18/runtime-config-client.html#GUC-LOCK-TIMEOUT)
- [PostgreSQL 18 Advisory Lock関数](https://www.postgresql.org/docs/18/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS)
- [PostgreSQL 18 pg_locks](https://www.postgresql.org/docs/18/view-pg-locks.html)
- [PostgreSQL 18 pg_stat_activity](https://www.postgresql.org/docs/18/monitoring-stats.html#MONITORING-PG-STAT-ACTIVITY-VIEW)
- [Node.js Release Schedule](https://nodejs.org/en/about/previous-releases)
- [Node.js v24.18.0 release](https://nodejs.org/en/blog/release/v24.18.0)
- [TypeScript 6.0 release notes](https://www.typescriptlang.org/docs/handbook/release-notes/typescript-6-0.html)
- [React versions](https://react.dev/versions)
- [React versioning policy](https://react.dev/community/versioning-policy)
- [Vite releases](https://vite.dev/releases)
- [Vite 8.2.0 release](https://github.com/vitejs/vite/releases/tag/v8.2.0)
- [PostgreSQL versioning policy](https://www.postgresql.org/support/versioning/)
- [PostgreSQL 18.4 release](https://www.postgresql.org/about/news/postgresql-184-1710-1614-1518-and-1423-released-3297/)
- [PostgreSQL 18 documentation](https://www.postgresql.org/docs/18/index.html)
- [npm workspaces](https://docs.npmjs.com/cli/v11/using-npm/workspaces/)
- [npm ci](https://docs.npmjs.com/cli/v11/commands/npm-ci/)
- [Vitest 4 guide](https://v4.vitest.dev/guide/)
- [Playwright release notes](https://playwright.dev/docs/release-notes)
- [Playwright browser management](https://playwright.dev/docs/browsers)
