---
name: frontend-e2e-test
description: topic2html のReact/Vite画面に対するPlaywright E2Eテストを設計・実装・レビューする。frontend/e2e、playwright.config.ts、fixture、feature別spec、Vite開発Server、Taskfile、traceや不安定なテストを扱うときに使用する。
---

# Frontend E2E Test

## 責務

- E2Eは、Browserから利用者がたどる重要な成功経路と、層をまたぐ重大な失敗だけを扱う。
- Storybookは部品・表示状態・操作・アクセシビリティ、Go単体テストはdomain/usecase/handlerの分岐を扱う。重複させない。

## 配置

```text
frontend/
├── playwright.config.ts
└── e2e/
    ├── fixtures/test.ts
    ├── setup/environment.teardown.ts
    └── <feature>/<user-flow>.spec.ts
```

- 全specは`fixtures/test.ts`から`test`と`expect`をimportする。
- 再利用されるlocator・操作だけを`*.page.ts`へ、秘密でない再利用データだけを`*.data.ts`へ切り出す。
- `test-results`、`playwright-report`、trace、videoをGit管理しない。

## 実装規則

- `webServer`でViteを起動し、既存Serverの再利用とCIの新規起動を区別する。
- role、label、accessible name、表示テキストを優先し、必要な場合だけ`data-testid`を使う。
- web-first assertionを使い、固定sleep、DOM階層、CSS class、XPathに依存しない。
- 各テストを単独・任意順で実行可能にする。実ユーザー設定、秘密情報、本番・共有DBへ接続しない。
- `forbidOnly`、CI時のretry、初回retryのtrace、最小限のChromium projectを明示する。

## 検証

```sh
task frontend:check
task frontend:build
task frontend:e2e
```
