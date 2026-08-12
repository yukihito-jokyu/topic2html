# topic2html E2E構成

- Playwright設定: `frontend/playwright.config.ts`
- テスト本体: `frontend/e2e`
- 共通入口: `frontend/e2e/fixtures/test.ts`
- 起動対象: Viteの開発Server
- Backend: Ginは`task dev`で同時起動できるが、E2EがHTTP APIを必要とする場合だけ明示的に接続する。
- 実行入口: `task frontend:e2e`

StorybookはUI状態、Go単体テストはdomain/usecase/handler、E2EはBrowserから確認する重要フローを担当する。実ユーザー設定、秘密値、本番・共有DBを使わない。
