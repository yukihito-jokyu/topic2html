---
name: frontend-implementation
description: topic2html のReact、TypeScript、Vite、Tailwind CSS、shadcn/uiフロントエンドを実装・改修する。frontend/src配下の画面、機能、コンポーネント、HTTP API境界、UI状態、アクセシビリティ、ビルド検証を伴う機能追加・修正・リファクタで使用する。
---

# Frontend Implementation

## 構造

```text
frontend/src/
├── app/            # アプリ全体の組立て
├── components/ui/  # shadcn/ui部品
├── features/       # 業務機能
├── hooks/          # 再利用する状態処理
└── lib/            # UI非依存の補助処理
```

`main.tsx`は起動だけに限定し、機能固有の表示・状態・HTTP呼出しは`features/<feature>`へ閉じ込める。共有UIは`components/ui`、汎用処理は`hooks`または`lib`へ置く。

## 実装手順

1. 利用者、操作、初期・読み込み・成功・空・失敗・再試行・操作中の状態を確認する。
2. `frontend/src`、`package.json`、Vite、TypeScript、Tailwind、`components.json`と既存部品を読み、既存の命名とimport aliasを使う。
3. HTTP通信が必要なら、画面から直接呼ばずfeature所属の薄いservice境界へ閉じ込める。Request/Responseの変換を画面へ漏らさない。
4. props、データ取得、表示、操作を分ける。小さな画面で不要な抽象化はしない。
5. キーボード、フォーカス、accessible name、disabled、長文、狭い幅を確認する。

## UI規則

- 新規の装飾用HTMLより、既存のshadcn/ui部品とそのvariantを優先する。
- Tailwindのsemantic tokenと`cn()`を使い、色の直接指定や条件classの文字列連結を避ける。
- リスト・表・詳細画面では派手さより走査性と操作効率を優先する。
- 外部依存は既存のReact、Radix、Tailwind系で足りない場合だけ追加し、理由と影響を確認する。

## 検証

```sh
task frontend:typecheck
task frontend:check
task frontend:build
task frontend:test
```

画面挙動が重要なら`task frontend:e2e`、部品や状態を追加したら`task frontend:storybook:test`も実行する。
プロジェクト共通の`task verify`がある場合は、単体テスト、Storybookテスト、E2Eテストを必要に応じてその入口へ含め、手元と継続的検査で検証範囲が分かれないようにする。
