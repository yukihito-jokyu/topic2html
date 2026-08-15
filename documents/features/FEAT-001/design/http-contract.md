# 管理認証HTTP契約

共通: 全応答に`Cache-Control: no-store`を付ける。JSON応答は`Content-Type: application/json; charset=utf-8`。認証失敗のJSONは`{"error":{"code":"unauthenticated"}}`、CSRF・Origin失敗は`{"error":{"code":"forbidden"}}`、入力不正は`{"error":{"code":"invalid_request"}}`、保護記録障害は`{"error":{"code":"authentication_unavailable"}}`とし、内部理由を含めない。Browser formの`oauth_start`は画面遷移を完結させる例外として、失敗理由をJSONで返さず固定失敗案内へ`303`する。保護記録を安全に読取り・更新できない場合、管理読取り・状態変更・logoutは`503 authentication_unavailable`で終了し、業務状態とsession cookieを変更しない。

Cookieは`__Host-topic2html_oauth_tx`と`__Host-topic2html_admin_session`を使用する。ともに`Secure; HttpOnly; SameSite=Lax; Path=/`、`Domain`未指定。前者の`Max-Age=600`、後者の`Max-Age=28800`。削除は同じ属性に`Max-Age=0`を追加する。CSRF header名は`X-CSRF-Token`、値はbase64urlの256 bit乱数である。

| operation | method / URI | 入力 | 成功 | 失敗 |
| --- | --- | --- | --- | --- |
| `oauth_start` | `POST /admin/auth/google/start` | Browser画面遷移として送る`application/x-www-form-urlencoded`。`Origin`必須かつtrusted app originと完全一致。任意field `return_path=/admin`、省略時`/admin`。現在の許可集合は`/admin`だけで、他の値は拒否。 | `303`、Google Authorization Endpointの`Location`、transaction cookie。`Location`のAuthorization Requestは`response_type=code`、環境専用client識別子、固定callback URI、`scope=openid email`、transaction固有の`state`と`nonce`、S256のPKCE `code_challenge`および`code_challenge_method=S256`を含む。Browserは`303`をdocument navigationとして追従する。 | Origin欠落・複数・不正、form fieldの重複・不正値、不正path、保護記録不可のいずれも、transaction cookieを発行・置換せず、`303 /admin/login?reason=failed`へ遷移する。`reason`は固定値のみで、Googleへredirectしない。 |
| `oauth_callback` | `GET /auth/google/callback` | query `code`と`state`。Google error時は`error`、任意`error_description`。transaction cookie必須。 | `303`で保存済み`return_path`へ、session cookie。 | 常にtransaction cookieを削除して`303 /admin/login?reason=failed`。`reason`は固定値のみ。 |
| `admin_session_bootstrap` | `GET /admin/auth/session` | session cookieのみ。 | `200 {"authenticated":true,"csrf_token":"..."}`。 | sessionなし・無効は`200 {"authenticated":false}`（該当cookie削除）。保護記録の読取りまたはCSRF ciphertext復号不可は`503 authentication_unavailable`。 |
| `admin_logout` | `POST /admin/auth/logout` | `Origin`必須。有効sessionなら`X-CSRF-Token`必須。 | `200 {"authenticated":false}`、session cookie削除。 | Origin不正は`403 forbidden`でcookie不変更。有効sessionのCSRF不正は`403 forbidden`でcookie不変更。 |

callbackはGoogle側で失敗してもBrowserへ技術詳細を返さない。`oauth_callback`の失敗redirectはsession cookieを発行しない。管理状態変更の共通ガードは後続FeatureのHTTP操作に、上記session cookie、`Origin`、`X-CSRF-Token`を適用する。有効session不正は`401 unauthenticated`（必要時cookie削除）、Origin/CSRF不正は`403 forbidden`、成功時のみ後続操作を実行しidle期限を更新する。

管理読取りの共通ガードは、後続Featureの管理プレビュー・履歴・任意版確認など、公開閲覧ではないread operationに上記session cookieを適用する。有効sessionなら読取りを通し、sessionなし・無効・期限切れ・許可メール不一致なら`401 unauthenticated`（必要時cookie削除）として業務データを返さない。読取りはCSRF tokenもidle期限更新も不要である。
