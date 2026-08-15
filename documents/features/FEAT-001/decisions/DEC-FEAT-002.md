# DEC-FEAT-002 — 管理認証の配置origin・Google OAuth登録運用

状態: **承認済み（2026-08-15、L3）**

## 課題

FEAT-001は、信頼済みアプリoriginへの厳格な`Origin`照合と、Google OAuthの固定redirect URIを必要とする。これらの値、Google Console登録、Secret管理の責任者が未確定では、安全な認可フローを実装・運用できない。

## 選択肢

### A. 環境ごとに固定originと専用Google OAuth Clientを登録する（推奨）

本番はHTTPSの単一の信頼済みアプリoriginを設定し、Google Consoleにはそのorigin配下の固定callback URIだけを登録する。ローカル開発・CIには本番と分離したOAuth Clientを使い、それぞれの許可originとcallback URIを個別に登録する。

各環境のorigin、callback URI、Client ID、SecretはServer運用設定として管理する。SecretはServer実行環境だけへ注入し、リポジトリ、Browser、生成HTML、ログ、fixtureには保存しない。配置を行う運用責任者がGoogle Console登録とSecret更新を担当し、値の不整合または欠落時はServerを起動させない。

### B. 複数originまたは共通OAuth Clientを柔軟に許可する

複数のoriginを同時に信頼し、同一OAuth Clientへ複数callback URIを登録する。

環境追加は容易になるが、CSRF許可範囲とGoogle Console登録の監査範囲が広がる。初期リリースの単一管理者・単一信頼済みアプリという前提に対して不要な運用複雑性を持ち込む。

## 推奨

**選択肢Aを採用する。** 環境境界を明確にし、Origin照合、redirect URI、Secret管理を一意に監査できるためである。

承認時には、少なくとも本番origin、固定callback URI、Google Console登録とSecret更新を担当する運用責任を明示する。本番の具体的なドメインが未取得の場合は、取得・登録をリリース前提条件として記録し、未設定のまま本番起動しない。

## 決定

**選択肢Aを採用する。** 本番の具体的なoriginとcallback URIは、配置前に運用責任者が確定・Google Consoleへ登録する環境設定とする。設定値の欠落または登録値との不一致では、本番OAuthログインを有効化しない。

## 承認後の影響

- Feature Designは、環境設定の論理契約、起動時検証、OAuth操作の固定callback利用、Origin照合、Secret非露出、Google test doubleの境界を具体化する。
- 本番origin・callback URI・Google登録が揃うまで、本番OAuthログインを有効化しない。
- PostgreSQL schema、migration、HTTP wire contract、transaction・cleanup、検証責務は、このDecisionの承認後に補助設計へ分離して確定する。

## 見直し条件

- 複数の信頼済み管理originを同時に提供するという、利用者・運用要件の変更が承認された場合。
- Google OAuth Clientの環境分離が、組織の統制または公式なGoogle運用制約と両立しないことが判明した場合。
