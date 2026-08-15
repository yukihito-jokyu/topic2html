# `admin_session_bootstrap` — 管理画面初期化

## 決定根拠

本資料は承認済みの[DEC-FEAT-001](../../decisions/DEC-FEAT-001.md)の選択肢Aを具体化する確定設計である。

## I/O

method、URI、JSON schema、status、cookie属性は[HTTP契約](../http-contract.md)に従う。

- 入力: 同一originの管理画面からの読取り要求とsession cookie。読取り要求ではsessionのアイドル期限を更新しない。
- 成功出力: 有効sessionのServer保護ciphertextを復号し、同一sessionのCSRF hashと一致を確認してから復元したCSRF tokenを`authenticated: true`とともに返す。許可メール、Google Token、session参照値、内部期限値を返さない。`Cache-Control: no-store`を必須とする。
- 匿名出力: `authenticated: false`。期限切れ・失効・不一致のsession cookieは削除する。`Cache-Control: no-store`を必須とする。
- 失敗出力: session記録を安全に読めない場合は、匿名と区別して`503 authentication_unavailable`を返し、管理操作を許可しない。cookieは削除しない。

## 制約

- 共通セッション契約に従い、管理認証操作は信頼済みアプリorigin以外へ`Access-Control-Allow-Origin`を返さず、資格情報を伴うcross-origin requestを許可しない。
- 有効sessionの判定は、sessionの存在、失効、アイドル期限、絶対期限、認可済みメールと現行許可メール設定の完全一致を満たすこととする。
- この操作は業務状態もsessionのアイドル期限も変更しない。ciphertextの復号または復号値hashと保存hashの定数時間比較に失敗した場合はtokenを返さず`503 authentication_unavailable`とする。
