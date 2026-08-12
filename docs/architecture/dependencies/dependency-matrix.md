# Module依存Matrix

- 対象作業: `L1-M1-S3`
- 状態: 設計済み
- 確認日: 2026-08-12

## 読み方

依存とは、あるModuleが別Moduleの公開された型、Port、DTO、Fixture、または実行入口を利用する関係である。Matrix（行と列の交点で関係を一覧する表）は「行のModuleが列のModuleへ依存してよいか」を示す。`○`は公開入口経由で許可、`—`は同一Moduleなので判定不要、`×`は直接依存を禁止、`条件`は後続Taskが契約を決めた後だけ公開入口経由で許可、を表す。

この表が必要なのは、Directoryが分かれていても逆向きのImportがあれば、FrontendからDatabaseへ到達したり、技術Adapterが業務状態を決めたりするためである。依存は必ず一方向にし、循環（AがBを参照し、Bも直接または間接にAを参照する状態）を禁止する。循環があると、変更順、初期化順、責任の所在を一意にできない。

## 許可・禁止Matrix

| 利用するModule ↓ / 利用されるModule → | Admin Frontend | Backend HTTP | Feature Application/Domain | Contract boundary | Codex Adapter | PostgreSQL Adapter | Migration Runner | Preview/Public Resolver | Credentialless Renderer | Test Support |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Admin Frontend | — | ○ | × | ○ | × | × | × | × | 条件 | ○ |
| Backend HTTP | × | — | ○ | ○ | × | × | × | 条件 | × | ○ |
| Feature Application/Domain | × | × | — | ○ | 条件（L3のPort経由） | ○（Repository Port経由） | × | 条件 | × | ○ |
| Contract boundary | × | × | × | — | × | × | × | × | × | ○ |
| Codex Adapter | × | × | ○（L3のPort実装のため） | ○ | — | × | × | × | × | ○ |
| PostgreSQL Adapter | × | × | ○（Repository Port実装のため） | ○ | × | — | × | × | × | ○ |
| Migration Runner | × | × | × | × | × | 条件 | — | × | × | ○ |
| Preview/Public Resolver | × | × | ○（L4/L6の公開Port経由） | ○ | × | × | × | — | ○ | ○ |
| Credentialless Renderer | × | × | × | × | × | × | × | × | — | ○ |
| Test Support | 条件 | 条件 | 条件 | ○ | 条件 | 条件 | 条件 | 条件 | 条件 | — |

## Matrixの補足

- Admin FrontendからRendererへの`条件`は、Rendererの実装・隔離方式を共有するという意味ではない。L6が定める不透明な表示導線だけを利用できる。管理DOMへの生成HTML挿入は許可しない。
- Backend HTTPからResolverへの`条件`は、RouteがResolverの用途別公開入口を呼ぶ場合だけである。Resolverを経由せずDatabase Recordを応答へ変換してはならない。
- FeatureからPostgreSQL Adapterへの`○`は、Featureが所有するRepository PortをAdapterが実装するための依存だけを表す。Featureがpgx、Pool、SQL、資格情報を直接利用することは許可しない。
- FeatureからCodex Adapterへの`条件`は、L3が生成用Portを定めた後だけである。Codex AdapterはL3のPort実装のためにFeatureへ依存できるが、HTTP、Frontend、Renderer、PostgreSQLの処理を呼び出さない。
- PostgreSQL AdapterからFeatureへの`○`は、Featureが定めたRepository Portを実装するためであり、Featureの業務状態遷移をAdapterが呼び出すことは許可しない。
- Migration RunnerからPostgreSQL Adapterへの`条件`は、後続のRunner実装で接続の公開境界を再利用する必要が承認された場合だけである。通常Backend起動、Feature、Frontendを経由してMigrationを実行しない。
- Test Supportの`条件`は、対象Moduleの公開された振る舞いを試験するためだけである。秘密値、Production環境、非公開内部状態をTestの前提にしない。

## 禁止方向と循環禁止の確認規則

1. `frontend/**`、`renderer/**`、`test/**`から`backend/adapters/postgresql/**`または`backend/migrations/**`へProduction依存を作らない。Database接続とMigration権限をBrowser側・表示面へ渡さないためである。
2. `backend/adapters/codex/**`、`backend/adapters/postgresql/**`と`backend/migrations/**`は`frontend/**`、`renderer/**`、`backend/http/**`を参照しない。技術的な外部接続、保存、Schema変更の処理が画面やHTTP入口を支配しないためである。
3. `contracts/**`は言語間の検査物であり、Feature Application/Domain、Adapter、HTTP Handlerの実装を参照しない。片方の実装詳細を契約の正本にしないためである。
4. `renderer/**`はResolverの公開入力だけを受け、管理Frontend、Backend HTTP、Feature、Adapterへ戻る依存を持たない。未信頼HTMLから管理領域へ戻る経路を構造上なくすためである。
5. 新しい依存を追加する前に、利用側OwnerはこのMatrixに既存の許可経路があるか確認する。ない場合は直接Importで解決せず、Feature OwnerとL7共有変更laneへ変更要求を渡す。Matrix変更はL1-M1-S3の設計変更として、後続Gateで再照合する。

このMatrixはModule境界を定めるものであり、具体的なGo Package import、TypeScript import、APIのField、RepositoryのQuery、Schemaを決めない。それらは各Feature Taskと後続の実装Taskが所有する。
