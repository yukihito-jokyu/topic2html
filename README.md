# topic2html

## 必要環境

- Node.js 24.18.0 / npm 11（Node.js版は `frontend/.nvmrc`）
- Go 1.26.5
- golangci-lint 2.12.2
- Task 3

## コマンド

```sh
task frontend:install
task frontend:storybook
task frontend:e2e
task frontend:e2e:report
task backend:coverage
task backend:integration
task db:create
task db:migrate
task dev
task verify
task run
```

Storybookは`task frontend:storybook`で起動し、`http://localhost:6006`で確認できます。
初回だけ`cd frontend && npx playwright install chromium`を実行してください。E2Eテストは`task frontend:e2e`で実行できます。一時PostgreSQL、Google OAuth test double、Backend、TLS Frontendを起動し、実アプリケーションのログイン・拒否・取消・logoutを検証します。実Googleや実資格情報は使用しません。実行ごとの録画は`frontend/test-results/`へ、HTMLレポートは`frontend/playwright-report/`へ保存されます（ともにGit管理外）。`task frontend:e2e:report`で最新レポートをブラウザーで開けます。CIでは録画とレポートを7日間、成果物としてダウンロードできます。

全コマンドは `Taskfile.yml` を参照してください。

`task backend:coverage` は、全Backend packageのstatement coverageが100%であることを検証し、`backend/coverage.html` を生成します。
`task backend:integration` は、一時PostgreSQLコンテナとダミーデータでrepository結合テストを実行し、終了時にコンテナを削除します。`TEST_DATABASE_URL` を指定すると、そのDBを使用します。
`task db:create` は開発用PostgreSQLを`127.0.0.1:5432`で起動します。`task db:migrate` はこのDBへmigrationを適用します。`task db:delete` はコンテナと全データを削除します。

### ローカル開発

初回は`.env.example`を`.env`へコピーし、Google Cloud Consoleで作成したローカル開発用のOAuth Client情報と、ログインを許可するメールアドレスを設定してください。`.env`はGit管理されず、Backendだけが読み込みます。シェル環境変数を設定済みの場合はそちらが優先されます。

```sh
cp .env.example .env
task dev
```

`task dev` は開発用PostgreSQL、migration、Backend（`http://127.0.0.1:8080`）、Frontend（`https://localhost:5173`）を起動します。初回は開発専用の自己署名証明書を`.certs/`へ生成します。Frontendの`/admin/auth`と`/auth`へのリクエストはBackendへproxyされます。終了は`Ctrl-C`です。

初回にブラウザで`https://localhost:5173`を開き、自己署名証明書を信頼してください。管理session cookieは`Secure`属性のため、このHTTPS Originを使う必要があります。Google Cloud Consoleのredirect URIは`https://localhost:5173/auth/google/callback`と完全一致させてください。実運用では`.env`や自己署名証明書を置かず、ホスティング環境のTLSとSecret管理から同じ環境変数を与えてください。

## 構成

- `backend/` は独立したGo moduleです。GinはHTTP adapterに限定し、実行入口、Domain、Application、外部adapterを分離します。
- `frontend/` は独立したNode.js/TypeScript作業領域です。Backend内部、PostgreSQL、Google、秘密情報には依存せず、将来の同一origin HTTP契約だけを利用します。
- Serverは `127.0.0.1:8080` で待ち受けます。認証設定が不正な場合は起動に失敗します。

設定値・Secretは環境からBackendだけへ渡し、リポジトリやBrowserへ置かないでください。
