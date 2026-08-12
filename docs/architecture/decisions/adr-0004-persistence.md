# ADR-0004: 永続化にPostgreSQLとpgxを採用する

- 状態: 採用
- 決定日: 2026-08-12
- 対象: 永続化

## 課題

本製品は、生成試行、形式検証済みHTML版、現在公開Pointer、タグ、配置、管理認可を区別して保存する。新版の保存だけでは現在公開版を変えず、公開切替や非公開化は明示操作として一貫して更新されなければならない。具体Tableをこの段階で決めずに、この関係と更新単位を支えられる保存技術とGo Backend向け接続手段を選ぶ必要がある。各状態の意味Ownerは別であり、特にL4は不変版、L6は現在公開Pointerを所有する。

## 評価基準

1. 関係、参照整合性、Transactionを扱えること。
2. 版と現在公開参照を別の状態として保存できること。
3. Schema Migrationと同時実行制御を標準的に扱えること。
4. GoからParameter化した明示Query、Connection Pool、Request中断を利用できること。
5. 現在の要件にない分散Database運用やObject-relational Mapperを前提にしないこと。

## 候補

- PostgreSQL + pgx v5: 採用する。PostgreSQLは関係制約とTransactionを持ち、pgxはPostgreSQL専用のGo DriverとToolkitを提供するため、版・公開参照・分類の整合とServer側接続を小さい構成で扱える。
- PostgreSQL + Go標準 `database/sql`: 採用しない。安定した共通Interfaceだが、本製品はPostgreSQLだけを対象とし、別Databaseへ差し替える要件がない。pgx公式もこの条件ではnative pgx Interfaceを推奨しており、不要な抽象化を増やさない。
- SQLite: 採用しない。小規模開始には簡単だが、Web BackendとWorkerが同時に保存へ触れる将来構成、Migration排他、運用上のBackup/監視を考えると、後からDatabase種別を移す費用が大きい。
- Document Database: 採用しない。HTML本文の保存自体は可能だが、版、公開Pointer、タグ、配置、試行の関係と一意条件をApplicationだけへ寄せる利点がない。
- Object Storageだけ: 採用しない。大きいHTML本文の格納候補にはなり得るが、状態遷移と関係のSystem of Record（正本）にはならない。

## 決定と採用理由

PostgreSQL 18系をSystem of Recordとして採用し、Go Backendからの接続へ `github.com/jackc/pgx/v5` の5.10系列を使う。`pgx` はGoとPostgreSQLの間でQuery、値変換、Transactionを扱う低水準Driver兼Toolkitである。Object-relational Mapper（`ORM`、Object操作からSQLを自動生成する層）は採用せず、状態遷移のSQLとTransaction境界を明示する。

同時Requestには `pgxpool` を使う。`pgxpool` は複数処理が再利用できる上限付きConnection Pool（Database接続を必要時に貸し出して戻す仕組み）であり、Requestごとの接続作成費用と無制限接続による資源枯渇を避けるために必要である。Poolの上限、待ち時間、接続寿命はIssue #7の環境別設定として扱い、Browserへ公開しない。

Queryへ利用者入力を文字列連結せず、`$1` などのPlaceholderとParameter値を別に渡す。Parameter化が必要なのは、入力をSQL構文として解釈させないためである。複数更新は明示的なTransactionにし、開始後は成功時に `Commit`、それ以外は `Rollback` する。pgxではContext取消だけでTransactionが自動Rollbackされないため、`defer` 等でRollbackを必ず呼べる制御を必要とする。

すべてのPool取得、Query、TransactionへHTTP Request由来の `context.Context`（中断・期限を伝えるGo標準の文脈）を渡す。期限後も接続待ちやQueryを続けないために必要である。ただし、Database資格情報をContext値へ保存・伝播させる用途は禁止する。

## 制約

- 許可: Server側RepositoryからのParameter化Query、明示的Transaction、上限付き `pgxpool`、Request Contextによる中断・期限伝播。
- 禁止: Browser/Rendererからの直接Database接続、Database資格情報のFrontend埋込み、文字列連結Query、起動時のORM自動Schema同期、Contextへの資格情報格納。
- 条件付き: HTML本文を将来Object Storageへ分離する場合も、版ID、状態、参照整合性の正本はPostgreSQLに残し、別ADRで整合方式を決める。
- Table、Column、Index、Repository、公開Pointerの具体表現、Backup/配備方式は本ADRの対象外である。
- PostgreSQL/pgx採用だけでは `RA-09` を保証しない。L2の認証・管理認可、L3の生成試行、L4の不変版、L5のタグ・配置、L6の公開状態は、各機能の意味を所有したままBackendの用途別Repository Port越しに永続化を利用する。Issue #6がPortとPostgreSQL AdapterのModule配置およびFrontend/Rendererからの依存禁止を決め、Issue #7がDatabase接続資格情報をBackendとMigration Runnerの外へ露出しない設定境界を決め、L7がFrontend成果物・Reference/CI実行トポロジーのNetwork経路・試験成果物を検査する。この分担が必要なのは、保存規則、物理依存、秘密情報、Reference/CIでの実行時検査のどれか一つだけではDatabaseと秘密情報への到達不能を保証できないためである。

## Version固定方針

- PostgreSQL 18.4を開始Baselineとする。2026-08-12時点の公式18.4 Releaseで確認できる現行18系Patchを起点とし、18系の後続PatchへはIntegration Test後に追随するが、Major 19へ自動更新しない。PostgreSQLの18.4はMajor 18内の修正Releaseであり、Schema版を意味しない。
- pgxは5.10.0を開始Baselineとし、`go.mod` に正確な版、`go.sum` にChecksumを固定する。v5は公式の最新安定Majorであり、5.10.0は確認日時点の公式Releaseである。変更し得るRepositoryの `master` だけではなく、v5.10.0 ReleaseとVersion指定したPackage文書を選定根拠にする。
- Database ServerとDriverの更新はMigration、Repository Integration、公開切替の同時更新Testを通す。

## 見直し条件

- PostgreSQL 18がSupport終了へ入る。
- 計測されたHTML容量や読取量により、Databaseだけでは運用目標を満たせない。
- 複数Regionでの書込み等、現在要件にない分散整合が必須になる。
- pgx v5が採用GoまたはPostgreSQL 18をSupportしなくなる。
- L4の不変版、またはL6の現在公開Pointerをそれぞれ安全なTransactionへできない。

## 公式一次資料

すべて2026-08-12確認: [PostgreSQL versioning policy](https://www.postgresql.org/support/versioning/)、[PostgreSQL 18.4 release](https://www.postgresql.org/about/news/postgresql-184-1710-1614-1518-and-1423-released-3297/)、[PostgreSQL 18 documentation](https://www.postgresql.org/docs/18/index.html)、[pgx v5.10.0 release](https://github.com/jackc/pgx/releases/tag/v5.10.0)、[pgx v5.10.0 package](https://pkg.go.dev/github.com/jackc/pgx/v5@v5.10.0)、[pgxpool v5.10.0 package](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool@v5.10.0)、[context package](https://pkg.go.dev/context@go1.26.5)。
