---
name: impl-react-admin-ui
description: topic2htmlのReact管理画面を実装する。承認済み画面仕様と同一originのHTTP contractに基づく画面状態、操作、アクセシビリティ、responsive対応で使用し、Server認可方式や業務画面の要件決定には使用しない。
---

# topic2html React管理画面実装

## Goal

Node.js/TypeScript、React、Viteでビルドする管理画面を、承認済みの画面仕様とHTTP contractに一致させて実装する。管理画面は信頼済みアプリoriginだけで動かし、秘密情報・許可メール・token・Cookie値を表示、URL、永続化、ログへ出さない。

## Use When

- 管理ログイン、生成、版・公開、タグ・掲載場所などの管理画面状態と利用者操作を実装・変更する。
- 同一originの管理APIを呼ぶReact UI、CSRF tokenの画面内保持、失敗状態、a11y、responsive対応を実装・変更する。
- 承認済みのscreen specificationを具体的な画面構成へ落とす。

## Do Not Use When

- HTTP、OAuth、session、CSRF、DB schema、外部連携のServer側contractを変更する（`impl-go-server-postgres`を使用し、必要なら設計へ差し戻す）。
- test double、統合/E2E基盤だけを構築する（`impl-project-verification`を使用）。
- 画面に存在しない業務フロー、文言方針、公開URIを新設する。

## Required Inputs

- 対象Featureの`implementation-handoff.yaml`、`tasks.md`、承認済み画面仕様
- 対象HTTP contract、operation、利用者操作列、テスト戦略
- `DEC-ARCH-001.md`、`DEC-ARCH-002.md`、`DEC-ARCH-003.md`と親`AGENTS.md`
- 親リポジトリの既存frontend構成、`package-lock.json`、UI規約

## Project Evidence

1. `DEC-ARCH-003`: 管理画面はNode.js 24.18.0/npm 11、TypeScript、React、Viteのビルド成果物であり、依存はnpm / `package-lock.json`で固定する。
2. `DEC-ARCH-001` / `DEC-ARCH-002`: 管理画面は信頼済みアプリoriginに属し、生成HTML隔離originへCookie、管理API client、認可情報を渡さない。
3. 親README: frontendのNode versionと統一Taskコマンドを管理する。
4. 実装コードはまだないため、初回の局所構造を一般規約へ過剰に昇格させない。

## Owned Implementation Scope

- 画面仕様の状態、表示、利用者操作、success / failure遷移
- 同一originの管理API clientと、responseから得る画面内CSRF tokenの短期保持
- form submission、document navigation、keyboard操作、フォーカス、読み上げ、responsive
- API失敗時の安全な画面復帰と再試行導線

## File Placement Rules

- 既存frontend構成があれば、それに従う。親READMEに示された`frontend/`のNode設定を尊重する。
- UI固有の表示・状態・styleは近接する画面またはcomponentの責務に置き、Server限定設定やSecretをfrontendへ置かない。
- browser storage、URL、telemetry、fixtureには認可token、Cookie値、許可メール、Secretを保存しない。
- 生成HTMLを実行・表示するUIはFEAT-005の承認済み隔離設計が整うまで追加しない。

## Autonomous Decisions

- 承認済み画面仕様を満たすcomponent分割、局所state、CSS、a11y実装
- 既存frontendの命名、styling、test patternへの追従
- JSON errorを利用者向けに安全な固定文言・状態へ対応付ける実装詳細

## Escalate When

- API responseにない状態、画面遷移、公開URL、利用者入力、業務操作が必要になる。
- CSRF tokenやCookieをbrowser storageへ保存する必要がある。
- 新しいUI framework・state管理・CSS libraryを採用する。
- screen specification、HTTP contract、実際のServer応答が矛盾する。

## Procedure

1. handoff、画面仕様、HTTP operation、利用者操作列を読み、対象状態と操作を列挙する。
2. 既存frontendの構成とtest規約を確認する。初期骨格だけなら未承認のデザインシステムやlibraryを導入しない。
3. 成功・失敗・未認証・基盤障害を含む画面状態と操作を実装する。
4. 非同期操作の途中状態、keyboard、focus、読み上げ、狭い表示域を確認する。
5. 秘密情報がDOM、URL、console、fixture、browser storageへ出ないことを確認する。
6. unit / component testと必要なbrowser E2Eを追加・更新してValidationを実行する。

## Validation

- `task frontend:install`（依存同期が必要な場合）
- `task verify`（親リポジトリの統一検証。利用可能な場合）
- TypeScript型検査、Biome整形検査・静的検査、production build、Vitest
- Google test doubleを用いる管理画面E2E

## Completion Criteria

- 画面仕様の状態、操作、API対応、a11y、responsive要件を満たす。
- API contract外の利用者挙動を暗黙に追加していない。
- 認可情報・秘密情報をClientへ保存または露出していない。
- 関連検証の結果と残る制約を報告する。
