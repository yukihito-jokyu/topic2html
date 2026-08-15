# 管理ログインから管理操作までの操作列

利用者は、許可済みGoogleメールの管理者である。目的は、CSRF tokenを安全に取得して管理読取り・状態変更・logoutを行うことである。

1. 管理画面はtrusted app originから、`return_path=/admin`を含むformで`POST /admin/auth/google/start`へ送る。Browserは`303`とtransaction cookieを受け、document navigationとしてGoogle認可画面へ遷移する。
2. Googleは`GET /auth/google/callback`へ戻す。Serverはtransactionを一回使用にし、検証・許可メール照合成功時にsession cookieと`303 /admin`を返す。失敗時はtransaction cookieを削除して固定失敗画面へ遷移する。
3. 管理画面は`GET /admin/auth/session`を呼ぶ。`{"authenticated":true,"csrf_token":"..."}`ならtokenを画面実行時メモリだけに保持する。`authenticated:false`ならログイン開始へ戻り、`503`なら一般化した再試行表示にする。
4. 管理読取りはsession cookieだけで共通read guardを通る。状態変更は同じcookieに加え、trusted `Origin`と`X-CSRF-Token`へ手順3のtokenを設定して共通mutation guardを通る。`401`は再ログイン、`403`はtoken再取得後の再試行、`503`は再試行表示とする。
5. logoutは`POST /admin/auth/logout`へtrusted `Origin`と、認可済みなら`X-CSRF-Token`を送る。`200 {"authenticated":false}`とsession cookie削除後、画面メモリのtokenも破棄する。`403`と`503`ではcookie・画面tokenを破棄せず、再試行または再初期化する。

全操作でURL、画面表示、ログ、fixtureにCSRF token、cookie値、OAuth値、許可メールを出さない。
