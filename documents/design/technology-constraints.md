# 技術制約

| 領域 | 状態 |
| --- | --- |
| 生成基盤 | 採用済み: Codex app-server |
| 本人確認 | 採用済み: Googleと許可メールアドレス |
| Backend Webランタイム・フレームワーク | 採用済み: Go 1.26.5、Gin v1.12.0 |
| Frontend | 採用済み: Node.js 24.18.0/npm 11、TypeScript、React、Vite |
| リポジトリ構造 | 採用済み: `backend/`と`frontend/`の分割、Backendの`handler`・`usecase`・`repository`・`domain`によるClean Architecture |
| 永続化ストア | 採用済み: PostgreSQL 18系、pgx/v5・pgxpool |
| 生成HTML隔離機構 | 方針採用済み: 別origin（具体方式は未決定） |
| 配備トポロジー | 未決定 |
| 実装基盤・運用検証 | 採用済み: DEC-ARCH-003（2026-08-15改定） |

## DEC-ARCH-003で採用した基盤

詳細は
[`DEC-ARCH-003`](decisions/DEC-ARCH-003.md) を正本とする。

- 信頼済みWebアプリは Go 1.26.5 と Gin v1.12.0、管理画面は Node.js/TypeScript のビルド成果物とする。
- ソース配置は`backend/`と`frontend/`へ分ける。Backendは`handler`・`usecase`・`repository`・`domain`を明示分割し、`cmd/`をcomposition rootに限定する。
- 関係データの正本は PostgreSQL とし、Goからは `pgx/v5` と `pgxpool` を使用する。
- Google本人確認は、サーバー側のOAuth 2.0 Authorization Code Flow と OpenID Connect の検証済みID Tokenに限る。
- Node依存はnpmと`package-lock.json`、Go依存（Ginを含む）はGo Modulesと`go.sum`で固定し、統一Taskから整形・静的検査・テストを実行する。
