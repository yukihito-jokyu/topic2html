# FEAT-001 画面設計仕様書 — 管理認証

## 対象と境界

本仕様は、信頼済みアプリorigin上の管理認証に必要な画面状態と操作を定める。対象は管理画面への入口、Google本人確認の開始・失敗後の再試行、認証済み状態の初期化、logoutである。

生成、版管理、公開、タグ、掲載場所の画面内容は対象外であり、各Featureが所有する。FEAT-001は、それらの管理画面が共通して利用する認証済み状態と操作可能条件だけを提供する。公開HTMLの閲覧画面はFEAT-005の責務である。

## 画面一覧

| 画面・状態       | 利用者・到達条件                                                        | 目的                                                                              | 退出先                                             |
| ---------------- | ----------------------------------------------------------------------- | --------------------------------------------------------------------------------- | -------------------------------------------------- |
| 管理画面初期化中 | `/admin`へ到達した利用者                                                | 現在の管理sessionを安全に確認する。                                               | 認証済み管理画面、ログイン案内、または再試行状態。 |
| ログイン案内     | sessionがない・無効、または`/admin/login?reason=failed`へ遷移した利用者 | Google本人確認を開始、または安全な失敗を再試行する。                              | Google認可画面、または`/admin`。                   |
| 認証済み管理画面 | 有効sessionが確認済みの管理者                                           | 後続Featureの管理機能を利用可能にする。FEAT-001はlogoutと認証状態だけを担当する。 | logout後のログイン案内。                           |
| 認証基盤障害     | session確認またはlogoutが`503`になった利用者                            | 秘密情報を出さず、再試行可能なことを伝える。                                      | 同一画面の再初期化またはlogout再試行。             |

Google callbackはBrowserのdocument navigationとして処理し、専用の管理画面を表示しない。成功時は`/admin`へ、失敗時は固定の`/admin/login?reason=failed`へ遷移する。

## 画面状態と操作

### 管理画面初期化中

- `/admin`の表示開始時に`GET /admin/auth/session`を一度呼ぶまで、後続Featureの管理読取り・状態変更操作は有効化しない。
- 読み込み中は、認証状態を確認中であることを表示する。管理データ、許可メール、token、cookie値、内部障害理由は表示しない。
- `200 {"authenticated":true,"csrf_token":"..."}`では、tokenを当該画面実行中のメモリだけに保持し、認証済み管理画面へ遷移する。
- `200 {"authenticated":false}`では、ログイン案内へ遷移する。tokenは保持しない。
- `503 authentication_unavailable`では、認証基盤障害を表示し、再試行操作だけを提供する。管理操作とlogoutを有効化しない。

### ログイン案内

- Googleで管理画面へログインする主操作を提供する。これは`return_path=/admin`を含む`application/x-www-form-urlencoded`のformを`POST /admin/auth/google/start`へ送る。JavaScriptのJSON requestや別origin requestでは送らない。
- 操作開始後は、Googleへの`303` document navigationが完了するまで同じ送信を重ねない。
- callback失敗で到達した場合は、本人確認に失敗したため管理操作を利用できないことだけを表示する。Googleの技術詳細、許可メール、token、`state`、`code`は表示しない。
- 失敗案内からの再試行は、同じGoogleログイン主操作を使う。利用者がログインを中断した場合にも、管理機能を部分的に表示しない。

### 認証済み管理画面

- 有効session確認後だけ、後続Featureが定める管理読取りと状態変更を表示・操作可能にする。
- 状態変更を行う画面操作は、保持中のCSRF tokenを`X-CSRF-Token`へ設定し、Browserが付与するtrusted `Origin`とsession cookieを前提にする。tokenをURL、画面表示、永続ブラウザ保存、ログへ出さない。
- 後続Featureの管理読取りが`401`なら、保持中tokenを破棄してログイン案内へ遷移する。状態変更が`403`なら、tokenを破棄して管理画面初期化から再取得する。`503`なら業務画面の状態を推測で更新せず、認証基盤障害として再試行を案内する。
- logout操作は認証済み状態でのみ表示する。`POST /admin/auth/logout`へ`Origin`と`X-CSRF-Token`を送る。成功したらtokenを直ちに破棄し、ログイン案内へ遷移する。`403`または`503`ではtokenを破棄せず、現在の状態を維持して再試行または再初期化を案内する。

### 認証基盤障害

- 表示する文言は、認証状態を確認・更新できないため管理操作を続行できないことと、再試行できることに限る。
- 再試行は`GET /admin/auth/session`を再実行する。logout中の`503`では`POST /admin/auth/logout`を再実行する。
- 秘密値、許可メール、DB・Google・通信の内部失敗理由を表示しない。

## API／operation対応

| 画面操作・状態     | operation                 | 入力                                       | 画面上の結果                                                                                            |
| ------------------ | ------------------------- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------- |
| `/admin`初期化     | `admin_session_bootstrap` | session cookie                             | 認証済み、ログイン案内、または認証基盤障害へ遷移する。                                                  |
| Googleログイン開始 | `oauth_start`             | formの`return_path=/admin`、trusted Origin | BrowserがGoogle認可画面へ遷移する。入力不正・Origin不正はログインを開始せず、安全な失敗案内を表示する。 |
| Google callback    | `oauth_callback`          | Googleのqueryとtransaction cookie          | 成功時は`/admin`、失敗時は固定失敗案内へBrowser遷移する。画面はOAuth値を扱わない。                      |
| 後続管理読取り     | `admin_read_guard`        | session cookie                             | 有効時だけ各Featureの読取り結果を表示する。`401`はログイン案内へ遷移する。                              |
| 後続管理状態変更   | `admin_mutation_guard`    | session cookie、trusted Origin、CSRF token | 成功時だけ各Featureが定める画面状態へ更新する。`401`／`403`／`503`は上記の共通動作に従う。              |
| logout             | `admin_logout`            | trusted Origin、有効session時のCSRF token  | `authenticated:false`でtokenを破棄しログイン案内へ遷移する。                                            |

operation別のHTTP・cookie・エラー契約は[HTTP契約](http-contract.md)を正本とする。画面はその契約を変更しない。

## アクセシビリティ・表示制約

- ログイン、再試行、logoutはキーボードだけで到達・実行でき、操作の目的が識別可能なラベルを持つ。
- 初期化・成功・失敗による画面状態の変化は、支援技術利用者が把握できるよう通知する。失敗時は再試行操作へ到達できる。
- 小さい画面幅でも主操作と失敗理由の安全な要約が、横方向スクロールなしに読めるよう再配置する。具体的な見た目・component構成は実装領域で決める。
- 許可メール、Google token、OAuth `state`／`nonce`／`code`、CSRF token、cookie参照値、Secret、内部障害詳細を画面・URL・支援技術向け通知に含めない。

## 検証対応

- E2Eで、未認証時のログイン案内、許可・不許可・取消後の遷移、認証済み時だけの管理操作、logout後のtoken破棄を確認する。
- HTTP契約テストで、画面操作がform POST、document navigation、cookie・CSRF契約を守ることを確認する。
- キーボード操作、状態通知、狭い画面幅で主操作が利用可能なこと、秘密値を表示しないことを確認する。
