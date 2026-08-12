---
name: storybook-impl
description: topic2html のReact/Vite UIに対してStorybookを実装・更新・診断する。frontend/.storybook、*.stories.tsx、Vitest addon、アクセシビリティ検査、UI状態、Storybook build・test、Taskfileの入口を扱うときに使用する。
---

# Storybook Implementation

## 作業前

`frontend/package.json`、lockfile、Vite、TypeScript、Biome、`src/styles.css`、`components.json`、`.storybook`、既存Story、Taskfileを読む。依存追加・更新時はStorybook、Vite、Vitestの公式互換要件を確認し、主要依存の更新が必要なら影響と選択肢を利用者へ確認する。

## 構成

- Storyは対象部品の近くに`*.stories.tsx`で置き、CSF 3、`Meta`、`StoryObj`、`satisfies`を使う。
- `preview.tsx`で`src/styles.css`、日本語locale、必要最小限のdecorator、a11y errorを構成する。
- Viteのaliasと解決規則を再利用し、Storybookだけに別のaliasを複製しない。
- `storybook-static`をGit・formatter・lint対象から除外する。

## 状態とテスト

- 通常、空、読み込み、失敗・再試行、無効化・操作中、長文、狭い幅から、見た目・振る舞い・アクセシビリティが異なる代表状態を選ぶ。
- Propsと`args`、callbackの`fn()`を優先する。HTTP通信はfeature service境界でモックし、実APIや実秘密情報へ接続しない。
- `play`ではrole、accessible name、label、利用者から観測できる表示・無効状態・フォーカスを検証する。固定sleep、CSS class、DOM内部構造に依存しない。
- DialogにはTitleを置き、操作経路、Escape、フォーカス復帰を確認する。

## 検証

```sh
task frontend:check
task frontend:build
task frontend:storybook:build
task frontend:storybook:test
```
