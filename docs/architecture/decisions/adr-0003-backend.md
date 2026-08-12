# ADR-0003: BackendにGo標準HTTP機能を採用する

- 状態: 採用
- 決定日: 2026-08-12
- 対象: Backend

## 課題

Backendは管理要求、公開対象解決、PostgreSQL、Codex app-serverをつなぐ。ただし、すべてを同じEndpoint（HTTPの個別入口）や応答Modelへ混ぜると、管理情報がRendererや匿名利用者へ漏れる。利用者がBackend言語として選んだGoを前提に、用途別の境界を実行時に検査でき、依存を増やし過ぎないHTTP構成が必要である。

## 評価基準

1. 公式に保守されるGo安定版で動作すること。
2. 管理、Preview解決、Public解決を別Route（HTTP MethodとURLで定まる入口）と別の要求・応答型にできること。
3. 外部JSON入力のMedia Type、有限Body長、単一JSON値、型、意味を実行時に検査し、応答Fieldを許可Listへ限定できること。
4. Requestの中断・期限をDatabaseや外部接続へ伝播できること。
5. Codex接続方式、認証方式、HTML形式検証方式を本ADRへ混ぜないこと。

## 候補

- Go標準Libraryの `net/http` と `http.ServeMux`: 採用する。`net/http` はHTTP Server/Clientの公式標準機能、`ServeMux` はMethod・Host・Pathに一致する処理へ要求を振り分ける標準Routerである。現在の用途別Routeに必要な機能を持ち、第三者Frameworkの更新・暗黙Bindingを増やさない。
- `chi`: 採用しない。標準 `net/http` と互換の軽量Routerで、Route GroupやMiddleware（複数Handlerに共通する認可・記録等を前後へ挟む処理）が増えた場合には有力だが、現要件では `ServeMux` で満たせる機能のために依存を一つ増やす理由がない。
- `Gin`: 採用しない。Binding、描画、MiddlewareをまとめたWeb Frameworkで実現可能だが、独自Contextや自動Bindingを採用する必要がなく、標準HTTP境界を直接読める構成より概念が増える。
- Node.js/Fastify: 採用しない。実現可能だが、利用者がBackendをGoに決定したため対象外である。Node.jsとTypeScriptは管理FrontendのBuild/Test専用として残し、Backend実行時には使わない。
- Full-stack React Framework内のServer Function: 採用しない。UIと管理/公開/Renderer境界がFrameworkの暗黙規約へ結合する。

## 決定と採用理由

BackendへGo 1.26.5を開始Baselineとして採用し、HTTP ServerとRouterは標準Libraryの `net/http` / `http.ServeMux` を使う。GoはCompileして単一の実行形式を作れる型付き言語であり、BackendのHTTP、業務処理、PostgreSQL接続、Codex接続をServer側に閉じるために使う。管理FrontendのTypeScriptとは言語を共有しないため、両者の契約は後述の正本Field定義と同一Fixture（同じ入出力例を機械試験へ渡す固定データ）で照合する。

各JSON Routeは用途別のGo構造体をRequest/Response Data Transfer Object（`DTO`、境界で許可したFieldだけを運ぶ型）として持つ。Decode前に `mime.ParseMediaType` で `Content-Type` をMedia Type本体とParameterへ分ける。欠落、構文不正、またはMedia Type本体が `application/json` 以外の入力は拒否する。Parameterは構文と値を検査し、省略または `charset=utf-8`（大文字小文字の差は無視）だけを許可し、未知ParameterとUTF-8以外は禁止する。`application/problem+json` 等の `+json` Media TypeはJSON文法を使っても別の契約を表すため、専用Route契約が将来定義されるまで受理しない。この限定が必要なのは、JSONとして読めることと、当該Routeの契約であることを混同しないためである。

Media Type検査後、`net/http.MaxBytesReader` でRequest Bodyの読取り上限を強制してから `encoding/json.Decoder` へ渡す。有限上限は大き過ぎる入力をDecodeやMemory消費の前に止めるために必要であり、全JSON Routeへ必ず設ける。各Route/APIの承認済み機能Ownerが、その契約で必要な最大入力からRoute固有のByte値を決め、Handlerへ適用して試験する。既に有限上限が重なる経路では、Route固有値をそれ以下へ厳格化することは許可するが、上限なしへ緩和してはならない。L7-M2-S1のComposition Owner（個別RouteをGlobal Routerへ組み付ける担当）は、Mount時に全JSON Routeへ有限cap（読取り量を打ち切る上限）があることを検査する。このADRは横断共通の既定Byte数も共通HTTP実装Ownerも新設しない。共有化が必要になった場合は、実装前に承認済みTask MapのOwner変更またはTask追加が必要である。L1-M2-S2はHTTP・Serializationを対象外とし、L1-M4-S1もHTTP/RPC Handlerを対象外とするため、どちらにもByte値やHandler適用を割り当てない。このADRでは機能Field、HTTP Status、Error値も先取りしない。

上限内のRequest JSONは標準 `encoding/json.Decoder` で構造体へ読み、未知Fieldを `DisallowUnknownFields` で拒否する。1回目のDecode後に2回目を行い `io.EOF`となることを確認し、JSON値の後ろに別の値が続く入力を拒否する。その後に必須性、長さ、列挙値、Field間条件を明示Validatorで検査する。構造体へのDecodeは「JSONの形と型」、Validatorは「製品上の意味」を検査する別責務であり、どちらか一方だけでは境界契約を満たさない。

Responseは用途別構造体からだけ `encoding/json` で生成し、Database Recordや内部Errorを直接Encodeしない。成功と失敗のどちらのJSON Responseも、BodyやStatusを書く前に `Content-Type: application/json` を設定する。応答の形をClientに明示し、Browserや中継層の別形式との誤解を防ぐためである。許可したFieldしか構造体に存在しないため、管理用元データが偶然匿名応答へ混ざるのを防ぐ。JSON Schemaの動的Compileは採用しないが、契約の実行時検査を省く意味ではない。L1-M4がField・型・条件の物理表現と共通Fixtureを決め、Go側ValidatorとFrontend側の型/検査が同じ契約を満たすことをContract Testで確認する。

各Handlerは `context.Context`（Requestの中断、期限、処理範囲の値を後続処理へ伝える標準の文脈）を受け渡す。利用者が接続を切った処理や期限超過処理をDatabase/Codexへ残さないために必要である。ただし、Contextへ資格情報や任意の業務データを無制限に詰める用途には使わない。

## 制約

- 許可: Server側のDatabase接続、認可済み管理処理、用途別Resolver、Codex app-server Adapter、用途別のGo Request/Response構造体と明示Validator、契約から物理化した有限Body上限。
- 禁止: 汎用Database RecordのHTTP返却、秘密情報のFrontend Build注入、Public/Rendererから管理Routeへの内部転送、未知JSON Fieldの黙認、Body上限なしのDecode、未定義 `+json` Media Typeの受理、利用者入力Schemaの動的実行。
- 条件付き: 管理RouteはL2の本人確認と管理認可の両方に合格した場合だけ利用できる。Public Resolverは認証を要求しないが、現在公開版以外を指定できない。
- `http.DefaultServeMux` のGlobal共有は使わず、明示的に作成した `http.NewServeMux` をBackend組立て時に注入する。Testと管理/公開Routeの構成を意図せず共有しないためである。
- Codex接続、Process、Prompt、Retry、失敗分類、HTML形式検証はL3の責任であり、本ADRはGo内部での具体方式を決めない。

## Version固定方針

- Go 1.26.5を開始Baselineとする。`go.mod` の `go` 行はこのModuleの最低必要Go版を1.26.5と宣言するが、実際に実行するToolchainの完全な固定ではない。そのためCIと開発手順に実行版1.26.5を別に明記する。公式Release Historyで確認できた2026-08-12時点の最新安定Patchを使う。
- 標準LibraryはGo本体と同じ版で固定されるため、`net/http` や `encoding/json` を別Versionとして取得しない。
- 第三者Moduleは `go.mod` に選定Version、`go.sum` に取得Moduleの暗号学的Checksum（同じ取得物か確かめる値）を記録し、`go mod verify` と全Test後に更新する。
- Go 1.27等のMinor更新は、Compile、Contract、Integration、Browser E2Eと標準Libraryの互換性確認後に取り込む。

## 見直し条件

- Go 1.26が公式Support終了へ入る。
- Route数やMiddleware要件が増え、標準 `ServeMux` では依存方向や一貫した境界検査を保てないことが実装・計測で判明する。
- Codex app-serverの公式InterfaceをGoから安定して利用できない。
- 要求量、Latency、長時間Jobの性質により現在のGo HTTP構成では運用要件を満たせないことが計測で判明する。
- 管理、Preview、PublicのRoute分離ではS1境界を満たせない。

## 公式一次資料

すべて2026-08-12確認: [Go Release History](https://go.dev/doc/devel/release)、[Go 1.26 release notes](https://go.dev/doc/go1.26)、[net/http package](https://pkg.go.dev/net/http@go1.26.5)、[MaxBytesReader](https://pkg.go.dev/net/http@go1.26.5#MaxBytesReader)、[mime.ParseMediaType](https://pkg.go.dev/mime@go1.26.5#ParseMediaType)、[encoding/json package](https://pkg.go.dev/encoding/json@go1.26.5)、[Go Modules Reference](https://go.dev/ref/mod)、[chi公式Repository](https://github.com/go-chi/chi)、[Gin公式Repository](https://github.com/gin-gonic/gin)。
