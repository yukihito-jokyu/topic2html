# 初期アーキテクチャ

信頼済みアプリケーション境界が認証、認可、コンテンツ状態、Codex app-server連携を所有する。生成HTMLは別の非信頼表示境界で実行し、アプリの権限を共有しない。詳細なAPI、画面、スキーマはFeature Designへ委譲する。

リポジトリは`backend/`と`frontend/`をトップレベルで分離する。`backend/`にはGo/Ginによる信頼済みWebアプリ、`frontend/`にはNode.js/TypeScriptによる管理画面だけを置く。生成HTMLの隔離originは、この分割とは独立した配置境界であり、管理画面の配信元やファイル配置によって信頼済みアプリoriginと混同しない。

Backendは、次の4層を明示したClean Architectureを採用する。`cmd/`（composition）は依存を組み立ててServerを起動するだけの外側であり、4層の業務責務を持たない。

```text
backend/
  cmd/          composition root（依存の組み立て・起動）
  handler/      HTTP入力・出力のadapter
  usecase/      アプリケーションユースケースと外部port
  repository/   PostgreSQL・Google・Codex app-server・設定などの外部I/O adapter
  domain/       業務概念、不変条件、業務規則
```

- `handler`はGinのrouting、middleware、HTTP request/response変換、認証済み主体の受け渡しを担う。ユースケースだけを呼び、SQL、外部SDK、環境変数、秘密値へ直接触れない。
- `usecase`はアプリケーション操作を表し、`domain`の規則を適用する。外部I/Oは自身が定義するport経由で要求し、Gin、PostgreSQL、Google、Codex app-server、環境変数には依存しない。
- `repository`は`usecase`のportを実装し、PostgreSQL、Google、Codex app-serverなどの外部詳細を閉じ込める。設定値は`cmd/`が読み取り・検証してconstructorから注入し、repository自身は環境変数を読まない。HTTP request/responseや業務フローの制御を持たない。
- `domain`は業務概念、不変条件、業務規則だけを持つ。ほかの3層、フレームワーク、HTTP、DB、外部provider、環境変数に依存しない。
- `cmd/`は設定を読み、repository実装、usecase、handlerを内側から外側へ接続する。業務規則、HTTP変換、外部通信の実装を置かない。

許可される依存は`handler → usecase → domain`、`repository → usecase/domain`、`cmd → handler/usecase/repository/domain`だけとする。`handler → repository`、`usecase → repository`、`domain → handler/usecase/repository/cmd`、およびrepositoryからhandlerへの依存は禁止する。HTTP request/response、SQL行、OAuth provider型、設定値を`domain`または`usecase`の公開モデルとして扱わない。

FrontendはBackendの内部パッケージやDBへ直接依存せず、同一originで公開されたHTTP契約だけを利用する。BackendはFrontendのビルド・画面実装へ依存しない。共通の業務規則を両者に複製しない。

承認済み: DEC-ARCH-001（単一の信頼済みWebアプリとリレーショナル永続化）、DEC-ARCH-002（生成HTMLを別originで隔離表示）、DEC-ARCH-003（Gin、Frontend/Backend分割、Clean Architecture）。
