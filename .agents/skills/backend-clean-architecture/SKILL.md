---
name: backend-clean-architecture
description: topic2html の Go/Ginバックエンドをクリーンアーキテクチャーで実装・改修・レビューする。internal/domain、internal/usecase、internal/repository、internal/handler/http、cmd/topic2html の責務分離、依存方向、HTTP契約、Repository境界、Goテストを伴う変更で使用する。
---

# Go/Gin Clean Architecture

## 構造

```text
backend/
├── cmd/topic2html/       # 設定・依存・HTTP Serverの組立て
└── internal/
    ├── config/           # 起動時設定
    ├── domain/           # 技術非依存の業務概念・不変条件
    ├── usecase/          # アプリケーション処理とRepository interface
    ├── repository/       # DB・外部API・ファイルなどの具体実装
    └── handler/http/     # GinのRequest/Response・Route
```

## 依存規則

- `domain` は Gin、HTTP、データベース、設定をimportしない。
- `usecase` は `domain` のみを使い、外部入出力に必要なinterfaceを所有する。
- `repository` は `usecase` のinterfaceを実装する。
- `handler/http` はGinの入力をusecaseの値へ変換し、domainまたはusecaseの結果をHTTP応答へ変換する。
- `cmd/topic2html` だけが具体repository、usecase、handlerを組み立てる。
- Ginの`Context`、HTTP status、JSON DTOをdomain/usecaseへ渡さない。取消・期限は標準の`context.Context`で渡す。

## 実装手順

1. 対象の業務概念と利用者から見える結果を確認する。
2. domainの値・不変条件、usecaseの入力・出力、Repository interfaceを内側から決める。
3. 外部接続の実装をrepositoryへ、HTTPの変換をhandlerへ置く。
4. `cmd/topic2html` で依存をコンストラクタ注入する。package変数や公開setterで注入しない。
5. handler、usecase、repositoryを責務別にテストする。HTTP契約はhandlerで、業務規則はusecaseで確認する。

## 禁止事項

- HTTP Request、Gin Context、JSON DTOをdomain/usecaseへ漏らさない。
- handlerからDB・外部APIを直接呼ばない。
- repositoryからGinやhandlerをimportしない。
- 将来用の空DTO、未使用interface、未接続repositoryを先取りしない。
- 秘密値、接続文字列、リクエスト全体をログに出さない。

## 検証

変更範囲に応じて、ルートで次を実行する。

```sh
task backend:format:check
task backend:lint
task backend:test
task backend:build
```
