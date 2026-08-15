# 技術制約

| 領域 | 状態 |
| --- | --- |
| 生成基盤 | 採用済み: Codex app-server |
| 本人確認 | 採用済み: Googleと許可メールアドレス |
| Webランタイム・フレームワーク | 未決定 |
| 永続化ストア | 未決定 |
| 生成HTML隔離機構 | 未決定 |
| 配備トポロジー | 未決定 |
| 実装基盤・運用検証 | 採用済み: DEC-ARCH-003（2026-08-14） |

## DEC-ARCH-003で採用した基盤

詳細は
[`DEC-ARCH-003`](decisions/DEC-ARCH-003.md) を正本とする。

- 信頼済みWebアプリは Go のHTTPサーバー、管理画面は Node.js/TypeScript のビルド成果物とする。
- 関係データの正本は PostgreSQL とし、Goからは `pgx/v5` と `pgxpool` を使用する。
- Google本人確認は、サーバー側のOAuth 2.0 Authorization Code Flow と OpenID Connect の検証済みID Tokenに限る。
- Node依存はnpmと`package-lock.json`、Go依存はGo Modulesと`go.sum`で固定し、統一Taskから整形・静的検査・テストを実行する。
