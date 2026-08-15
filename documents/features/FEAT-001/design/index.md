# FEAT-001 補助設計資料索引

本索引は、[詳細設計](../design.md)を補う実装準備済みの契約資料の入口である。方式は承認済みの[DEC-FEAT-001](../decisions/DEC-FEAT-001.md)および[DEC-FEAT-002](../decisions/DEC-FEAT-002.md)に従う。

| 資料 | 内容 |
| --- | --- |
| [Clean Architecture境界](architecture-boundaries.md) | `backend/`と`frontend/`の責務、`cmd`/`handler`/`usecase`/`repository`/`domain`の依存方向、移行・構造検証。 |
| [共通セッション契約](session-contract.md) | OAuth transaction、管理session、cookie、CSRF、共通の失敗・秘密情報保護。 |
| [PostgreSQL schema・migration](database-schema.md) | 保護記録のDDL相当契約、migration、DB access map。 |
| [HTTP契約](http-contract.md) | operation別method、URI、入力、status、JSON、header、cookie。 |
| [設定・外部境界](runtime-configuration.md) | origin、callback、Secret、起動時検証、Google境界。 |
| [検証責務](test-strategy.md) | unit、integration、HTTP、E2E、境界テストの分担。 |
| [管理ログイン・操作列](use-cases/admin-login-and-management.md) | token受渡しを含むログイン、管理操作、logoutの利用者手順。 |
| [画面設計仕様書](screen-specification.md) | 管理認証に必要な画面、状態、操作、API対応、アクセシビリティ。 |
| [oauth-start](operations/oauth-start.md) | Google認可開始。 |
| [oauth-callback](operations/oauth-callback.md) | Google callback、ID Token検証、管理session作成。 |
| [admin-session-bootstrap](operations/admin-session-bootstrap.md) | 同一originの管理画面を初期化する読取り操作。 |
| [admin-mutation-guard](operations/admin-mutation-guard.md) | 全管理状態変更の認可・CSRFガード。 |
| [admin-read-guard](operations/admin-read-guard.md) | 全管理読取りのsession認可ガード。 |
| [admin-logout](operations/admin-logout.md) | 管理sessionの失効。 |
