# 管理認証の設定・外部境界契約

## Server限定設定

| 論理設定 | 値・検証 | 所有・非露出 |
| --- | --- | --- |
| trusted app origin | 環境ごとに一つのabsolute origin。本番はHTTPS、path/query/fragmentなし。 | 配置運用責任者。Browserへ設定値として配らない。 |
| OAuth callback URI | trusted app originと固定path`/auth/google/callback`から構成し、Google Console登録値と完全一致。 | 同上。 |
| Google Client ID / Secret | 環境専用OAuth Clientの値。 | SecretはServer実行環境だけへ注入。Client IDはOAuth Authorization Endpointの`Location` queryに必要な公開識別子としてだけ現れ、JSON応答・ログ・fixtureへは出さない。 |
| 許可メール | 検証済みメールと完全一致で比較する単一値。 | Server限定。 |
| DB接続情報・保護鍵 | PostgreSQL接続とPKCE verifier暗号化に必要な値。 | Server限定。鍵は環境ごとに分離する。 |

起動時に全設定の存在・形式、origin/callbackの整合、Secretと保護鍵の空値でないことを検証する。いずれかが不正なら管理認証endpointを提供せず起動を失敗させる。本番の具体値とGoogle Console登録はリリース前提条件であり、配置運用責任者が変更を同期する。

## Google境界

Google Authorization Endpoint、Token Endpoint、OIDC discovery/JWKSはServerからだけ利用する。Token交換とJWKS取得には10秒のtimeoutを設定し、認証フロー中にretryしない。一時的障害、非2xx、JSON不正、署名・claim検証不能は全て認証失敗として扱う。通常CI/E2Eは固定のdiscovery、JWKS、Token応答を返すtest doubleを使用し、実Google資格情報へ接続しない。
