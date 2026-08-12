# ADR-0006: Frontend npmとBackend Go ModulesでBuildを再現する

- 状態: 採用
- 決定日: 2026-08-12
- 対象: Build

## 課題

管理Frontend、Go Backend、境界契約、Migration、Testの成果物を、開発者とCIで同じToolchain（Sourceから検査・Build・Testする道具一式）と依存版から再現する必要がある。FrontendとBackendは別言語になったため、Node.js/npmだけへBackendを含めず、二つの再現可能な経路を明示しなければならない。一方、まだModule/Directory設計前なので、複雑なMonorepo Orchestratorや配備方式を先に導入しない。

## 評価基準

1. FrontendのNode.js/npm依存とBackendのGo Module依存を正確に再現できること。
2. TypeScript型検査、Vite Build、Go Compileを明示Commandにできること。
3. Frontend/Backend間の契約不一致を、言語共有に頼らず機械検査できること。
4. 追加Serviceや大規模Task Orchestratorなしで開始できること。
5. 後続Issue #6のModule境界と、未選定の配備方式を制約しすぎないこと。

## 候補

- npm + Go Modulesの二Toolchain: 採用する。FrontendはNode.js同梱npmと `package-lock.json`、BackendはGo公式Module機構と `go.mod` / `go.sum` で、それぞれの正規手段を使える。
- Repository全体をnpm Workspacesだけで管理: 採用しない。Frontend Packageには使えるが、Go SourceとGo依存をnpm Packageとして扱うとBackendの標準Toolchainを隠す。
- Bazel/Nx等で全言語Buildを統合: 採用しない。大規模Task Graph、Remote Cache、専用規則は現在のModule数とTeam規模には過剰である。
- Docker ImageだけをBuild定義とする: 採用しない。配備再現には有効だが、言語ごとの型検査・依存検証・Testを置き換えず、Container製品と配備基盤も未選定である。

## 決定と採用理由

FrontendはNode.js 24.18.0と同梱npm 11を使い、正確なJavaScript依存を `package-lock.json` へ固定する。`npm ci` はLockfileとManifestが一致しないと失敗し、既存依存Directoryを作り直してLockfileどおり取得するCI向けCommandである。npm Workspace（Repository内の複数JavaScript Packageをまとめて扱う機能）は、Issue #6がFrontend/Testに複数Packageを置くと決めた場合だけ使えるが、Go BackendをWorkspaceへ含めない。

BackendはGo 1.26.5とGo Modulesを使う。`go.mod` はModule名、利用Go版、直接・間接依存Versionを宣言するManifest、`go.sum` は取得Module内容のChecksumを記録する検証台帳である。`go mod download` で依存を取得し、`go mod verify` でModule Cacheの取得物が記録と一致することを確認し、`go build` でBackend実行形式を作る。Go標準LibraryはGo本体に含まれるため、npmや別Lockfileでは管理しない。

二言語間のAPI契約は「TypeScript型をGoへ直接共有」しない。L1-M2-S2が定義する論理契約を、L1-M4が言語非依存のField定義とCanonical Fixture（正本として両実装が読む固定JSON例）へ物理化し、Frontendの型/検査とGoの構造体/Validatorが同じFixtureへ合格するContract Testで照合する。この仕組みが必要なのは、どちらか一方のCompile成功だけでは相手言語とのRuntime互換を証明できないためである。具体的な契約形式とCode生成の採否はIssue #6がModule配置と共に決め、本ADRはJSON Schema採用を先取りしない。

最低限のBuild入口を、FrontendのInstall・Type Check・Vite Production Build、BackendのModule検証・Compile、両者のTestとして分ける。Command名やDirectoryはIssue #6が決めるが、片方の成功で全体成功としない。

## 制約

- 許可: Frontend用npm/`package-lock.json`、Backend用Go Modules/`go.sum`、言語別の明示Command、CIのClean取得と検証。
- 禁止: Lockfile/`go.sum` なしの未固定取得、CIで固定Fileを書き換えるInstall、Go Backendをnpm Workspace Packageとして扱うこと、Frontend Buildへの秘密値埋込み、未使用Orchestrator導入。
- 条件付き: npm WorkspacesはIssue #6が複数Frontend/Test Packageを必要とした場合だけ使う。Build CacheやContainer Buildは、計測された時間または配備要件が生じた場合だけ、決定責任を持つOwnerを明示した後続ADRで選ぶ。
- Container BuildはApplication Processと依存物を配備Image（同じ実行内容を再作成する読取専用のひな型）へまとめる工程であり、Browser内のHTML表示枠を作る工程ではない。Container製品と本番Hosting製品は本ADRおよび現在のL7範囲では選定しない。
- Build成功はS1の隔離保証ではない。Frontend成果物へ秘密や管理Clientが混入しない検査はL7で別に行う。
- Lint/Format Toolは本Taskでは選定しない。L1-M3で必要性を確認し、採用する場合は正確な版と規則を固定する。

## Version固定方針

- Frontend ToolchainはNode.js 24.18.0/npm 11、Backend ToolchainはGo 1.26.5を開始Baselineとする。Backendの `go.mod` は最低必要Go版を宣言し、CIと開発手順は実際に使うGo 1.26.5を固定する。この区別は、依存先の要求版とBuildの実行版を同じものと誤解しないために必要である。
- npm依存の正確なVersionとIntegrity Hash（取得物が同一か確認する値）は `package-lock.json`、Go ModuleのVersionとChecksumは `go.mod` / `go.sum` へ固定する。
- Node 24系Patch/npm 11系またはGo 1.26系Patch更新は、Clean取得、型/Module検証、全Build、全Test後に取り込む。
- 次のMajor/Minor Toolchainへ自動移行せず、全ADRのSupport、契約Fixture、成果物差分を確認する。

## 見直し条件

- 二Toolchainの実行時間が計測上の開発阻害となり、共通Task RunnerやCacheが必要になる。
- Issue #6でCanonical Fixtureだけでは契約差分を十分に検査できず、言語間Code生成が必要になる。
- 配備基盤が再現可能なContainer/Image Buildを必須にする。
- Node.js 24/npm 11またはGo 1.26がSupport終了となる。
- Native依存などにより同じ対象環境向け成果物を再現できない。

## 公式一次資料

すべて2026-08-12確認: [Node.js Release Schedule](https://nodejs.org/en/about/previous-releases)、[Node.js v24.18.0](https://nodejs.org/en/blog/release/v24.18.0)、[npm workspaces](https://docs.npmjs.com/cli/v11/using-npm/workspaces/)、[npm ci](https://docs.npmjs.com/cli/v11/commands/npm-ci/)、[Go Release History](https://go.dev/doc/devel/release)、[Go Modules Reference](https://go.dev/ref/mod)、[go.mod reference](https://go.dev/doc/modules/gomod-ref)、[go command documentation](https://go.dev/cmd/go/)。
