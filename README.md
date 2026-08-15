# topic2html

## 必要環境

- Node.js 24.18.0 / npm 11（Node.js版は `frontend/.nvmrc`）
- Go 1.26.5
- Task 3

## コマンド

```sh
task frontend:install
task frontend:storybook
task backend:coverage
task backend:integration
task db:create
task db:migrate
task verify
task run
```

Storybookは`task frontend:storybook`で起動し、`http://localhost:6006`で確認できます。
初回だけ`cd frontend && npx playwright install chromium`を実行してください。E2Eテストは`npm run test:e2e`で実行できます。

全コマンドは `Taskfile.yml` を参照してください。

`task backend:coverage` は、全Backend packageのstatement coverageが100%であることを検証し、`backend/coverage.html` を生成します。
`task backend:integration` は、一時PostgreSQLコンテナとダミーデータでrepository結合テストを実行し、終了時にコンテナを削除します。`TEST_DATABASE_URL` を指定すると、そのDBを使用します。
`task db:create` は開発用PostgreSQLを`127.0.0.1:5432`で起動します。`task db:migrate` はこのDBへmigrationを適用します。`task db:delete` はコンテナと全データを削除します。

## 構成

- `backend/` は独立したGo moduleです。GinはHTTP adapterに限定し、実行入口、Domain、Application、外部adapterを分離します。
- `frontend/` は独立したNode.js/TypeScript作業領域です。Backend内部、PostgreSQL、Google、秘密情報には依存せず、将来の同一origin HTTP契約だけを利用します。
- Serverは `127.0.0.1:8080` で待ち受けます。認証設定が不正な場合は起動に失敗します。

OAuth/OIDC、PostgreSQL migration、session/CSRF、管理画面は後続Taskで実装します。設定値・Secretは環境からBackendだけへ渡し、リポジトリやBrowserへ置かないでください。
