---
name: backend-integration-test
description: topic2html のGo/Ginバックエンド結合テストを設計・実装・レビューする。usecaseをAPI境界とし、Docker Composeの専用DB、Goのseed、build tag、Taskfile、テストデータ、DB差分を扱うときに使用する。
---

# Backend Integration Test

## 方針

- 結合テストの主境界は`internal/usecase`の公開メソッドとする。Gin handlerはHTTP契約テストで扱う。
- 本番・個人・共有DBを使わず、Docker Composeで起動する専用DBと、リポジトリ管理のGo seedだけを使う。
- 純粋な業務規則の重複ではなく、usecaseから観測できるDB制約、方言差、外部I/Oを検証する。

## 配置

```text
backend/internal/usecase/
  <feature>_integration_test.go
backend/test/integration/
  db/
  seed/
```

- 結合テストには`//go:build integration`を付ける。
- 本番コードから`test/integration`をimportしない。
- 接続先は環境変数で渡し、資格情報やDSNをテスト名・ログ・成果物へ出さない。

## Taskfile

DBを導入するTaskでだけ、次を追加する。

```text
integration:up    専用DBを起動する
integration:test  go test -tags=integration ./... を実行する
integration:down  専用DBを停止する
```

通常の`task backend:test`へ結合テストを混ぜない。各テストまたはsubtestが独立して再実行できるよう、作成したデータを確実に片付ける。
