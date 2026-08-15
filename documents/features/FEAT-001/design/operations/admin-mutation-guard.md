# `admin_mutation_guard` — 管理状態変更の認可・CSRFガード

## 決定根拠

本資料は承認済みの[DEC-FEAT-001](../../decisions/DEC-FEAT-001.md)の選択肢Aを具体化する確定設計である。

## 役割

後続Featureが追加する全ての生成・版管理・公開・タグ・掲載場所の**状態変更**の直前に適用する共通ガードである。成功時だけ後続Featureの業務処理を開始できる。失敗時は業務状態を一切変更しない。`admin_logout`の匿名・期限切れ・失効済みsessionにおけるcookie削除は、業務状態を変更しない明示的な例外であり、同操作資料の契約を優先する。

## 必須検査とI/O

適用するcookie、CSRF header、401/403 JSONは[HTTP契約](../http-contract.md)に従う。

次を順に全て満たす必要がある。

1. 要求の宛先が信頼済みアプリoriginである。
2. session cookieがあり、対応するsessionが存在し、失効しておらず、アイドル期限・絶対期限内である。
3. sessionの認可済みメールが、現行のServer限定許可メール設定と完全一致する。
4. `Origin` headerが信頼済みアプリoriginと完全一致する。`Origin`が欠落・複数・不正なら拒否する。`Referer`は補助ログには使えても、CSRF判定の代替にしてはならない。
5. 管理画面が専用headerで送るCSRF tokenが存在し、sessionに記録したハッシュと一致する。

- 成功出力: 後続の状態変更操作へ通過する。全検査を通過した状態変更要求だけ、sessionのアイドル期限を延長できる。
- session不正: `401`を返し、業務状態を変更しない。必要に応じてsession cookieを削除する。ただし`admin_logout`は、信頼済みoriginの検査に成功し、sessionが匿名・期限切れ・失効済みと判明した場合に限り、この応答を使わず、cookie削除と`authenticated: false`を返す。
- CSRF不正: `403`を返し、業務状態を変更しない。許可メール、token、内部検証理由は返さない。

cookieを添付するだけのGET/HEADは業務状態を変更してはならない。これにより、OAuth callback以外のクロスサイトGETによる状態変更を禁止する。
