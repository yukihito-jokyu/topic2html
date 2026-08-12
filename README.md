# topic2html

## 必要環境

- Node.js 24.18.0 / npm 11（Node.js版は `frontend/.nvmrc`）
- Go 1.26.5
- Task 3

## コマンド

```sh
task frontend:install
task verify
task frontend:dev
task run
task dev
```

全コマンドは `Taskfile.yml` を参照してください。

## 初期骨格の制約

- Serverは `127.0.0.1:8080` のみで待ち受けます。
- 認証、Database、外部サービス、秘密値は扱いません。
