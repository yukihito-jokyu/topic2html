# FEAT-001 詳細設計レビュー

## 対象と独立性

`features/FEAT-001/{requirements.md,design.md,design/**,decisions/**}`を、設計者および過去のreviewerとは異なるfresh reviewerとして、正規要件、Initial Design、承認済みDecisionへ反証的に照合した。今回追加された[画面設計仕様書](design/screen-specification.md)と、Browser form POSTの`oauth_start`失敗時に固定失敗案内へ`303`遷移する契約を重点確認した。

## 正規ソース・Decision台帳

| 優先度 | ソース | 状態 | 適用範囲・確定内容 |
| --- | --- | --- | --- |
| 1 | `requirements/requirements.md` REQ-026–027 | confirmed | 事前登録済みGoogleメールアドレスだけが管理操作を行える。 |
| 1 | `requirements/business-rules.md` BR-001、BR-015–016 | confirmed | 管理者は許可メール1件だけであり、管理情報・秘密情報を公開・失敗表示へ出さない。 |
| 1 | `requirements/nonfunctional-requirements.md` NFR-001、NFR-003 | confirmed | 認証情報、session、管理操作・管理情報を生成HTMLへ渡さない。 |
| 2 | `design/decisions/DEC-ARCH-001.md`、`DEC-ARCH-002.md` | 承認済み | 信頼済みWebアプリが認証・認可を所有し、生成HTMLは別originで隔離する。 |
| 2 | `design/decisions/DEC-ARCH-003.md` | 承認済み L3 | React管理画面、Go Server、PostgreSQL、Server側OAuth/OIDC検証、Google test doubleを採用する。 |
| 2 | `features/FEAT-001/decisions/DEC-FEAT-001.md` | 承認済み L3 | Server側永続session、同期CSRF token、状態変更時の厳格なOrigin照合を採用する。 |
| 2 | `features/FEAT-001/decisions/DEC-FEAT-002.md` | 承認済み L3 | 環境ごとの単一trusted app origin、固定callback、環境専用OAuth Clientを採用する。 |
| 3 | `planning/traceability.md`、`design/{architecture,cross-cutting-concerns}.md` | 確定 | REQ-026–027はFEAT-001であり、業務画面は後続Feature、隔離表示はFEAT-005の責務である。 |

## 契約根拠の追跡

| 境界 | 設計上の契約 | 根拠 | 判定 | レビュー結果 |
| --- | --- | --- | --- | --- |
| 本人確認・認可 | Server側OIDC検証済みメールを許可メールと完全一致で照合する。 | REQ-026–027、BR-001、DEC-ARCH-003 | explicit | callback、session、設定資料で整合する。 |
| session・CSRF | opaque session、同期CSRF token、状態変更時のtrusted Origin照合を必須にする。 | DEC-FEAT-001 | explicit | HTTP・operation・画面仕様で整合する。 |
| 配置・外部OAuth境界 | 環境ごとの単一trusted app originと固定callback URIをServer設定から構成する。 | DEC-FEAT-002、DEC-ARCH-003 | explicit | 設定・Secret非露出・Google test doubleの契約は整合する。 |
| 管理読取り・匿名公開 | 管理読取りは有効sessionを要し、公開HTMLの匿名閲覧には認証を要求しない。 | REQ-026–027、BR-015、DEC-ARCH-001 | derived | read guardと画面仕様はFEAT-005の公開閲覧責務を侵食していない。 |
| 管理認証画面の範囲 | `/admin`の認証初期化、ログイン開始、callback後の遷移、認証状態・logoutだけを扱う。 | FEAT-001要件、traceability、DEC-ARCH-003 | derived | 画面仕様は業務画面の詳細を後続Featureへ残している。 |
| OAuth開始の画面遷移 | 同一originのBrowser form POSTは成功時にGoogleへ、入力・Origin・保護記録の失敗時に固定失敗案内へ、いずれも`303` document navigationする。 | HTTP契約、`oauth-start` operation | derived | 画面仕様の失敗案内への到達経路と整合する。 |
| 秘密情報非露出・画面アクセス性 | token、cookie値、許可メール、OAuth値、Secret、内部理由を出さず、キーボード操作・状態通知・狭い画面幅を考慮する。 | BR-015–016、DEC-ARCH-003 | derived | 画面仕様は根拠のある詳細化で、未承認のプロダクト判断を導入していない。 |

## 資料間整合監査

- 要件→Feature設計→画面設計仕様書→HTTP／operation資料を照合した。画面仕様は、対象利用者、到達・退出、初期化・ログイン案内・認証済み・認証基盤障害の状態、主要操作、`200`／`401`／`403`／`503`の画面動作、API対応、アクセシビリティ、responsive、秘密情報非露出を追跡可能にしている。
- 画面仕様は、FEAT-001が所有する認証初期化・Googleログイン・callback後の遷移・logout・共通認証済み条件だけを定める。生成・版管理・公開・タグ・掲載場所の画面内容を定めず、FEAT-005の公開HTML閲覧にも介入しない。Feature境界の逸脱は確認されなかった。
- `oauth_start`は同一originのform POSTだけを受ける。成功時はtransaction cookie設定後にGoogle Authorization Endpointへ`303`し、Origin欠落・不正form・不正復帰先・保護記録不可時はtransaction cookieを発行・置換せず固定の`/admin/login?reason=failed`へ`303`する。画面仕様はこの遷移で安全な失敗案内を表示し、再試行も同じformで行うため、旧来のJSON応答と画面表示の矛盾は解消されている。
- callback成功時の`303 /admin`、失敗時のtransaction cookie削除と固定失敗案内への`303`、専用callback画面を表示しないことは、HTTP契約、`oauth_callback` operation、画面仕様で一致する。`reason`は固定値であり、OAuth値・許可メール・内部理由をURLまたは画面へ渡さない。
- `/admin`初期化、`GET /admin/auth/session`の`200`／`503`、認証済み後のread／mutation guard、logoutのCSRF token・`403`／`503`の挙動は、HTTP契約、operation資料、操作列、画面仕様で一致する。tokenをURL・表示・永続ブラウザ保存へ出さない制約も整合する。
- 画面の外観、component構成、file path、symbolは実装領域へ残されている。新たな公開契約、権限、保存先、運用責任を確定する未承認L3/L4 Decisionは検出されなかった。

## 差し戻し事項

なし。

## 総合判定

**pass**

画面設計仕様書はFEAT-001の認証境界を守り、HTTP・operation・test資料と整合する。特にBrowser form POSTの`oauth_start`失敗時は、JSONを文書表示する不整合を残さず、固定失敗案内への`303` document navigationにより安全な再試行経路が成立している。未承認L3/L4 Decision、秘密情報露出、後続Featureの画面責務の混入は確認されなかった。

## 次ゲート

独立した`implementation-readiness-review`を実行する。`pass`後も利用者による詳細設計の明示承認前にTask分割・implementation handoffへ進んではならない。
