# 管理認証の設定・外部境界契約

## Server限定設定

`cmd`だけが環境変数・Secret storeから全Server設定を読取り、存在・形式、origin/callback整合、Secret・保護鍵の非空値を起動時に検証する。検証に失敗した場合は管理認証endpointを提供せず起動を失敗させる。`cmd`は、認可に必要な検証済みの非秘密依存を`usecase`へ、Client Secret・保護鍵・DB接続情報をGoogle/保護記録`repository`へconstructor注入する。`repository`は環境変数・Secret storeを直接読まない。Gin `handler`、`domain`、Frontendも直接読まない。`usecase`の公開入出力には一般的な環境設定モデルやSecretを含めない。

| 論理設定 | 値・検証 | 所有・非露出 |
| --- | --- | --- |
| trusted app origin | 環境ごとに一つのabsolute origin。path/query/fragment/userinfoなし。本番はHTTPS。非本番ではHTTPS、または`localhost`、IPv4 loopback、IPv6 loopbackだけをhostとするHTTPを許可する。HTTPを非loopback hostへ設定してはならない。 | 配置運用責任者。Browserへ設定値として配らない。 |
| OAuth callback URI | trusted app originと固定path`/auth/google/callback`から構成し、Google Console登録値と完全一致。 | 同上。 |
| Google Client ID / Secret | 環境専用OAuth Clientの値。 | SecretはServer実行環境だけへ注入。Client IDはOAuth Authorization Endpointの`Location` queryに必要な公開識別子としてだけ現れ、JSON応答・ログ・fixtureへは出さない。 |
| 許可メール | 検証済みメールと完全一致で比較する単一値。 | Server限定。 |
| DB接続情報・保護鍵 | PostgreSQL接続とPKCE verifier暗号化に必要な値。 | Server限定。鍵は環境ごとに分離する。 |

起動時に全設定の存在・形式、origin/callbackの整合、Secretと保護鍵の空値でないことを検証する。いずれかが不正なら管理認証endpointを提供せず起動を失敗させる。本番の具体値とGoogle Console登録はリリース前提条件であり、配置運用責任者が変更を同期する。

loopback HTTPの許容はローカル開発・限定的な非本番接続の設定検証だけを緩和するものであり、cookie属性を緩和しない。`__Host-`接頭辞と`Secure`属性は全環境で維持する。したがって、Browserでtransaction/session cookieを往復させるローカルまたはCIのOAuth E2Eは、TLSを持つloopback trusted app originを使用し、test doubleのcallbackもその固定HTTPS URIへ戻す。HTTP loopbackを使う場合は、Secure cookieを要するBrowser OAuth E2Eの成功条件に含めない。

## Google境界

Google Authorization Endpoint、Token Endpoint、OIDC discovery/JWKSは`backend/repository/`のGoogle実装からだけ利用する。Authorization Requestのscopeは`openid email`に固定し、`email_verified`を含むID Tokenのメールclaimを得るために必要な最小scope以外を要求しない。`usecase`はGoogle通信・SDK・HTTP型を知らず、検証済み本人情報または安全な失敗分類をport経由で受ける。Token交換とJWKS取得には10秒のtimeoutを設定し、認証フロー中にretryしない。一時的障害、非2xx、JSON不正、署名・claim検証不能は全て認証失敗として扱う。通常CI/E2Eは固定のdiscovery、JWKS、Token応答を返すtest doubleを使用し、実Google資格情報へ接続しない。
