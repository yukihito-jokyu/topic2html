# DEC-FEAT-001 — 管理認可セッション・CSRF・Google OAuth操作

状態: **承認済み（2026-08-14、L3）**

## 課題

FEAT-001では、許可メールと一致するGoogle本人確認済み利用者だけに管理操作を許可する。DEC-ARCH-003はGoogle OAuth/OIDCのServer側検証と`state`、`nonce`、PKCE、ID Token検証を承認済みとした。管理ログイン状態の継続と、状態変更へのCSRF防御の方式を本Decisionで確定する。

## 選択肢

### A. Server側の永続sessionと同期CSRF tokenを使う（推奨）

Server側の永続session記録と、値自体に意味を持たない短命・`HttpOnly` cookieを組み合わせる。管理画面が状態変更する際は、Server発行の同期CSRF tokenと`Origin`照合の両方を必須にする。

権限失効・強制ログアウト・OAuth transactionの一回使用をServerで確実に制御でき、cookie窃取以外のクロスサイト要求を二重に防げる。DBへのsession記録と失効処理が必要である。

### B. 自己完結cookieと二重送信CSRF tokenを使う

署名付き自己完結cookieだけでログイン状態を持ち、CSRF tokenをcookieと要求値の二重送信で照合する。

DB書込みは減るが、即時失効、session一回使用、Server側のセキュリティ監査を扱いにくい。tokenをJavaScriptから読めるcookieにする必要もある。

## 決定

**選択肢Aを採用する。** 管理者が1名でも、生成・公開・管理記録を扱うため、認可失効と異常時のfail closedをServer側で一貫して管理する価値が大きい。

## 承認後の影響

選択肢Aの承認により、以下の操作設計を適用し、FEAT-001のTask分割はそれらを前提にしてよい。

- [共通セッション契約](../design/session-contract.md): Server保護記録を正本とするopaque管理sessionと短期OAuth transaction、8時間の絶対期限、30分のアイドル期限、10分のtransaction期限と一回使用。
- [操作資料索引](../design/index.md): `HttpOnly` session/transaction cookie、session内の同期CSRF token、状態変更時の厳格な`Origin`照合、認可・CSRFガードとfail closed。

具体的なDB table名・列名、HTTP URI、migrationは後続のFeature Design補助資料で確定する。画面コンポーネント、file path、symbolは、このDecisionを変えない範囲で実装領域に委ねる。

## 見直し条件

- Server側の永続session記録によって、承認済みの単一信頼済みWebアプリとリレーショナル永続化の境界を満たせないことが、実装検証で判明した場合。
- `Origin`照合と同期CSRF tokenの併用では、同一originの管理画面から安全に状態変更できないことが判明した場合。
- Google OAuth/OIDCまたは配置originの承認済み前提が変更される場合。
