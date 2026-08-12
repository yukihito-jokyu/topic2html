---
name: go-test-style
description: topic2html のGoテストを作成・修正・レビューする。testing.Tの失敗処理、期待値比較、サブテスト、context、HTTP handler、internal配下のGoテストを扱うときに使用する。
---

# Go Test Style

## 規則

- 後続がpanicする、または前提が崩れる場合だけ`Fatal`/`Fatalf`を使う。独立して比較できる値は`Error`/`Errorf`でまとめて報告する。
- 失敗表示は`got`と`want`を両方示す。
- 複数ケースでは、入力と期待値を明確に表せるテーブル駆動テストと` t.Run`を使う。単一ケースを無理に表にしない。
- `context.Context`を受ける処理にはnilを渡さない。HTTP requestは`httptest.NewRequestWithContext`を使う。
- goroutine内から`Fatal`/`Fatalf`を呼ばない。失敗をchannelで親goroutineへ返す。
- 新しい責務のテストは対応する`<name>_test.go`に置き、実装の詳細ではなく公開された振る舞いを検証する。

## クリーンアーキテクチャー別の確認

- domain/usecaseはGin、HTTP status、JSON DTOを知らない状態で検証する。
- handlerはGin Routerを通じ、HTTP statusとJSON契約を検証する。
- repositoryの外部I/Oは単体テストと結合テストを分ける。

## 検証

```sh
task backend:format:check
task backend:lint
task backend:test
```
