# `admin_logout` — 管理session失効

## 決定根拠

本資料は承認済みの[DEC-FEAT-001](../../decisions/DEC-FEAT-001.md)の選択肢Aを具体化する確定設計である。

## I/O・状態変更

method、URI、JSON schema、status、header、cookie属性は[HTTP契約](../http-contract.md)に従う。

- 入力: 信頼済みアプリoriginと完全一致する`Origin`を含む要求。認可済みsessionを失効する場合は、正しいCSRF tokenも必須とする。
- 成功出力: sessionの失効とsession cookie削除後に、`authenticated: false`を返す。
- 既に匿名・期限切れ・失効済みの場合: `Origin`検査に成功したときだけ、session cookieを削除して`authenticated: false`を返す。これは共通ガードの`401`の例外であり、業務状態を変更せず、既存sessionを復活させない。CSRF tokenは要求しない。
- CSRF検査に失敗した場合: sessionを失効させず、`403`を返す。利用者は同一originの管理画面で再試行できる。
- `Origin`検査に失敗した場合: sessionの有無にかかわらず、cookieを削除せず`403`を返す。

有効sessionでは[admin-mutation-guard](admin-mutation-guard.md)の全認可・CSRF検査を適用する。匿名・期限切れ・失効済みsessionでは、同ガードのsession不正`401`を適用せず、上記の明示的な冪等cookie削除分岐を適用する。
