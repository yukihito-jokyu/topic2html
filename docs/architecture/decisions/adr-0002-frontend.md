# ADR-0002: 管理FrontendにReactとTypeScriptを採用する

- 状態: 採用
- 決定日: 2026-08-12
- 対象: Frontend

## 課題

生成、Preview、版履歴、公開切替、タグ、配置という管理操作には、非同期処理と複数の状態がある。管理画面を部品化しつつ、生成HTMLの実行面を管理画面から明確に分けられるFrontend技術が必要である。

## 評価基準

1. TypeScriptでGo Backendの言語に依存しないAPI契約と整合を検査できること。
2. 管理画面の状態と部品の責務を分けられること。
3. 生成HTMLをReact管理DOMへ挿入しない構成を自然に保てること。
4. Server Side Rendering（`SSR`、ServerでHTMLを組み立てる方式）を必須にしないこと。
5. 専用Frameworkの規約を過剰に導入しないこと。

## 候補

- React: 採用。状態と画面部品の組合せに集中したLibraryで、実行境界やRouting方式を独自に決められる。
- Vue: 採用しない。実現可能だが、Reactに対するこの製品固有の優位性がなく、候補を二つ維持する理由がない。
- Next.js等のFull-stack Framework: 採用しない。`SSR`、Server/Client Component、Data Cacheなど本製品に必須でない実行概念を増やし、管理Frontendと公開Rendererの境界をFramework規約へ結び付ける。
- Vanilla DOM API: 採用しない。依存は減るが、複数の非同期状態、画面部品、試験境界を手作業で維持する負担が大きい。

## 決定と採用理由

管理FrontendへReact 19.2系列とTypeScript 6.0系列を採用する。Browser向け開発配信とBuildにはVite 8.2系列を使う。確認日時点で通常の不具合修正を受ける現行Minorが8.2であり、一つ前の8.1は重要修正とSecurity修正のBackport対象に限られるため、新規Baselineを旧Minorへ留めない。

TypeScriptはFrontend専用であり、Go BackendとSource型を直接共有しない。L1-M4が物理化する言語非依存のField定義とCanonical Fixture（正本としてFrontend/Backend双方の試験が読む固定JSON例）に対し、Frontendの型とRuntime検査、Goの構造体とValidatorが同じ結果になることをContract Testで確認する。この区別が必要なのは、TypeScriptのCompile成功だけではGoが実際に受け取るJSONの必須性・許容値まで保証できないためである。

Reactは管理画面専用とする。生成HTMLはReact Component、React State、管理DOMの子要素にせず、ADR-0001のRendererへNavigationまたは隔離された表示面として渡す。この限定が必要なのは、UI Libraryの便利な共有機構から `RA-06`・`RA-07` の禁止対象へ到達する経路を作らないためである。

## 制約

- 許可: 管理画面のComponent、管理用Routing、管理API Client、画面状態管理。
- 禁止: 生成HTML文字列の管理DOMへの挿入、Rendererへの管理Context/Callback/API Clientの受渡し、Browser Bundleへの秘密情報注入。
- 条件付き: Build時に公開してよい設定値はIssue #7で公開区分を決めた値だけである。秘密値は名前を変えてもFrontendへ含めてはならない。
- Reactは管理UIの選定であり、Rendererの具体的隔離方式、機能UI、Accessibility指標を決めるものではない。

## Version固定方針

- BaselineはReact 19.2、TypeScript 6.0、Vite 8.2とする。
- 実装開始時の正確なPatch Versionを `package-lock.json` へ固定し、`npm ci` で再現する。
- ReactはSemantic Versioning（互換性の種類をMajor・Minor・Patchで表す版規則）に従う安定版だけを使う。
- Viteは公式方針上TypeScript型がMinor Versionで変わり得るため、Minor更新でも型検査・Build・Browser Testを必須にする。
- TypeScriptは6.0から7への移行期にあるため、7系へ自動更新しない。

## 見直し条件

- React 19またはVite 8がSupport対象外になる。
- L6のRenderer分離をReact管理画面から独立して構成できない。
- 要件上SSRが必要になり、Client-only管理画面では満たせない。
- FrontendのTypeScript契約とGo Backendの実行時検査の乖離をCanonical Fixtureで自動検出できない。

## 公式一次資料

2026-08-12確認: [React versions](https://react.dev/versions)、[React versioning policy](https://react.dev/community/versioning-policy)、[Vite releases](https://vite.dev/releases)、[Vite 8.2.0 release](https://github.com/vitejs/vite/releases/tag/v8.2.0)、[Vite 8 announcement](https://vite.dev/blog/announcing-vite8)、[TypeScript 6.0 release notes](https://www.typescriptlang.org/docs/handbook/release-notes/typescript-6-0.html)。
