# topic2html Storybook構成

- frontend: React 19、TypeScript 6、Vite 8
- styling: Tailwind CSS 3、shadcn/ui、`frontend/src/styles.css`
- alias: `@/`は`frontend/src`
- Storybook: `frontend/.storybook`、部品に近接する`*.stories.tsx`
- 検証: `task frontend:check`、`task frontend:build`、`task frontend:storybook:build`、`task frontend:storybook:test`

HTTP通信が必要な画面はfeature service境界でモックする。実行中のGin Server、外部API、秘密値へStoryから接続しない。
