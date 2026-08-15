# FEAT-001 要件

## 目的

事前登録されたGoogleメールアドレスの利用者だけに、管理操作を許可する。

## 対象

- REQ-026、REQ-027
- BR-001、BR-015
- NFR-001、NFR-003（管理者資格を信頼済みアプリ境界に留める制約として参照）
- DEC-ARCH-003（Ginを`handler`に限定し、`backend/`と`frontend/`を分離した、`handler`/`usecase`/`repository`/`domain`のBackend Clean Architecture）

## 対象外

- 生成、版管理、公開、タグ、掲載場所
- 閲覧者向け公開HTMLの表示
- Google本人確認プロトコルおよび画面技術の選定
- 生成HTMLの別origin隔離、表示ポリシー、隔離の結合検証（FEAT-005の責務）
- FEAT-002以降が所有する生成、版、公開、タグ、掲載場所の`domain`/`usecase`設計
