# Module・Directory境界

- 対象作業: `L1-M1-S3`
- 状態: 設計済み
- 確認日: 2026-08-12

## この文書の目的

この文書は、後続の実装がどの責務をどのModule（変更理由が同じ処理をまとめ、外部から使える入口を限定する単位）へ置くかを定める。Directory（Fileを置く場所）は、その境界をRepository上で確認できるようにする配置である。まだ実装RepositoryにはSource Directoryがないため、ここで示すPathは将来作る**論理Path**、すなわち責務とOwnerを示す予定の場所であり、File名、HTTP API、Table・Column、型、具体的なデータモデルを決めるものではない。

境界が必要なのは、Browserへ渡してよいものとServer側だけに残すもの、機能固有の意味とPostgreSQL等の技術詳細を分け、禁止された到達経路をSource依存の段階で作れないようにするためである。技術の採用理由は[横断技術構成](../technology-stack/overview.md)、永続化の制約は[ADR-0004](../decisions/adr-0004-persistence.md)、二Toolchainと契約Fixtureの制約は[ADR-0006](../decisions/adr-0006-build.md)を正本とする。

## 構成図

```text
Browser
  ├─ Admin Frontend ──> Go Backend HTTP ──> Feature Application/Domain ──> Repository Port
  │                         │                         │                         └─> PostgreSQL Adapter
  │                         │                         └─> Codex Adapter（L3が具体化）
  │                         └─> Contract boundary <──> Contract Fixture/Test
  └─ Credentialless Renderer <── Preview/Public Resolver <── Feature Application/Domain

Migration Runner ───────────────────────────────────────────────────────> PostgreSQL
Test Support ──> 各公開された境界（Frontend、Backend、Contract、Renderer、Migration）
```

矢印は「左側が右側の公開された入口を利用する」ことを表す。矢印のない方向の直接参照は許可しない。特に、Rendererは管理Frontendの一部ではなく、資格情報を持たない別の表示面である。Migration Runnerも通常Backend起動の一部ではなく、Schema変更だけを明示的に実行する別の入口である。

## Moduleと論理Directory

| Moduleと論理Path | 責務 | 許可する利用先 | 禁止事項と必要な理由 |
| --- | --- | --- | --- |
| Admin Frontend: `frontend/**` | 管理者がBrowserで操作する画面、管理用DTOとAPI Clientを持つ。 | Backend HTTPが公開する管理用契約、Rendererへ渡す不透明な導線。 | PostgreSQL、Codex、秘密設定、生成HTMLのDOMを直接参照しない。Browserは利用者に観測されるため、秘密や未信頼HTMLを同じ実行面へ置けない。 |
| Go Backend HTTP: `backend/http/**` | HTTP Routeの入口で、用途別Handler、Media Type、Route固有の有限Body上限、丁度一つのJSON値、応答形式を検査する。 | 機能Applicationの公開入口、Go側DTO/Validator、Contract boundary。 | Database Recordを直接応答せず、上限のないDecode、未定義`+json`、未信頼Schemaの動的実行をしない。外部入力と内部保存形式を分けるためである。共通HTTP実装や具体的なByte値のOwnerはここで新設しない。 |
| Feature Application/Domain: `backend/features/**` | L2〜L6の各機能Ownerが所有する業務意味、状態遷移、用途別Portを置く。 | 自分のFeature内のDomain、Repository Port、必要なAdapterの公開入口。 | 他Featureの内部実装、Frontend、Renderer、PostgreSQL Driverを直接参照しない。機能の意味を技術詳細や別機能の都合で変えないためである。具体的な状態・Schemaは各Feature Taskが決める。 |
| Contract boundary: `contracts/**` | 言語非依存Field定義、Canonical Fixture、Go/TypeScriptのContract Testを置く。Canonical Fixtureは両言語が同じように受理・拒否できるかを確認する正本JSON例である。 | Frontend側の型/検査、Backend側DTO/Validator、Contract Test。 | GoまたはTypeScriptの一方だけの型を正本にせず、Featureの業務実装を入れない。二言語のCompile成功だけでは実行時のJSON互換性を保証できないためである。形式とCode生成の採否は後続Taskの承認対象である。 |
| Codex app-server Adapter: `backend/adapters/codex/**` | L3が所有する生成要求を、Server側からCodex app-serverへ接続する。 | L3のFeatureが公開するPort、Codex app-server。 | Frontend、Renderer、PostgreSQL Adapterを直接参照しない。生成用接続と資格情報をBrowser側や保存処理へ混ぜず、`CON-001`の接続先固定を守るためである。接続・Process・Prompt・再試行の具体方式はL3が決める。 |
| PostgreSQL Adapter: `backend/adapters/postgresql/**` | FeatureごとのRepository Portをpgx/pgxpoolによるParameter化Query、Transaction、Pool利用へ接続する。 | Featureが公開するRepository Port、PostgreSQL Driver。 | Browser、Admin Frontend、Rendererから直接参照されない。Table・Column・Index・Repositoryの具体表現をここで決めない。永続化技術をFeature意味から分離し、資格情報とDatabase到達をServer側へ閉じるためである。 |
| Migration Runner: `backend/migrations/**` | Version付きSQLを単一の明示入口から実行し、Migration専用Session、有限Timeout、排他、失敗時停止、接続解放を担当する。 | PostgreSQLとMigration Toolだけ。 | 通常Backend起動で自動実行せず、Backendの業務Handler、Frontend、Rendererから参照しない。稼働中のApplicationが意図せずSchemaを変えないためである。Migrationの論理統治はL1-M2-S3、Runner実装はL1-M4-S2が所有する。 |
| Preview/Public Resolver: `backend/resolver/**` | Previewと匿名Publicで異なる対象選択を行い、Rendererへ解決済みHTMLと列挙済み最小表示情報だけを渡す。 | L4の不変版Content Port、L6の公開Pointer/Capability、Renderer公開入口。 | Rendererから管理機能へ戻るPort、汎用Record/Errorの返却、任意版の匿名解決を許可しない。Preview/Publicの選択差と同じ隔離強度を両立させるためである。具体契約はL6が決める。 |
| Credentialless Renderer: `renderer/**` | 未信頼の生成HTMLを、管理情報を持たない表示面で表示する。 | Resolverが列挙した最小表示情報、L6が決める表示契約。 | 管理API Client、認証状態、Database、Codex、管理Storage、管理Frontend DOMを参照しない。生成HTMLから管理領域や秘密情報へ戻る経路を作らないためである。 |
| Test Support: `test/**` | Go Test、Frontend Test、Contract Test、PostgreSQL Test Instance、Browser E2Eの補助を置く。 | 対象Moduleの公開境界、Test専用の代用品・Fixture。 | Production秘密情報と、実装Moduleの非公開内部状態に依存しない。実際の利用経路を検査し、通常CIから秘密を出さないためである。 |

## 境界を越えるときの規則

1. Module間の利用は、相手Moduleが公開したPort、DTO、Fixture、または明示された入口だけを通す。内部DirectoryのImportは、後から内部変更を安全にできなくなるため禁止する。
2. FeatureはPostgreSQL Adapterを直接の意味Ownerにしない。Featureが用途別Repository Portを定義し、AdapterはそのPortを技術へ接続する。これにより、保存技術の変更が業務規則を勝手に変えない。
3. FrontendとBackendは別Toolchainである。Frontendのnpm/`package-lock.json`とBackendのGo Modules/`go.mod`・`go.sum`を混ぜず、両者の境界はContract Fixture/Testで照合する。これはADR-0006の制約である。
4. 具体的なDirectoryの作成、Package名、Route、Field、Table、Column、Index、Code生成、Container/Hostingはこの設計の対象外である。必要になった場合は、それを所有する後続TaskまたはADRで決める。
