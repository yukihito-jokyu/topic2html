# 管理認可の共通セッション契約

## 決定根拠

本資料は、承認済みの[DEC-FEAT-001](../decisions/DEC-FEAT-001.md)の選択肢Aを具体化する確定設計である。Google OAuth/OIDCのServer側検証、`state`、`nonce`、PKCE、ID Token検証はDEC-ARCH-003で承認済みの不変前提である。

## 共通状態

| 概念 | 保持場所 | 内容・制約 |
| --- | --- | --- |
| OAuth transaction | Serverの短期保護記録 | `state`のハッシュ、`nonce`、PKCE verifier、発行時刻、有効期限、使用済み状態、固定の復帰先を持つ。一回限りで開始から10分で期限切れとする。PKCE verifierはServer復号可能な保護領域にのみ保持する。 |
| transaction cookie | Browser → 信頼済みアプリorigin | 意味を持たないランダムなtransaction参照値だけを保持する。`Secure`、`HttpOnly`、`SameSite=Lax`、`Path=/`、`Domain`未指定とし、callback後・失敗後・期限切れ後に削除する。 |
| 管理session | Serverの永続記録 | 意味を持たないランダムなsession参照値のハッシュ、認可済みメール、発行時刻、最終利用時刻、絶対期限、失効時刻を持つ。各管理要求で現行の許可メール設定とも完全一致を再確認する。 |
| session cookie | Browser → 信頼済みアプリorigin | session参照値だけを保持する。cookie名は`__Host-`接頭辞、`Secure`、`HttpOnly`、`SameSite=Lax`、`Path=/`、`Domain`未指定とする。生成HTMLの隔離originには送られない。 |
| CSRF token | Server保護された管理session記録と管理画面の実行時メモリ | Serverがsessionごとに生成する推測困難な値。Serverは照合用hashと、Server保護鍵で暗号化した復元用ciphertextだけを記録し、同一originの管理画面へ安全な初期化応答で値を渡す。cookieには入れず、状態変更要求の専用headerで送る。 |

乱数値は少なくとも256 bitの暗号学的に安全な乱数を用い、比較はタイミング攻撃を避ける。DBにはcookie値、`state`、CSRF token、PKCE verifierの平文を保存しない。

管理sessionの絶対有効期限は8時間、アイドル期限は30分とする。認可・CSRF検査を通過した状態変更要求だけがアイドル期限を延長でき、絶対期限は延長しない。読取り要求は延長しない。期限切れ・失効・許可メール設定との不一致では、Serverは認可を失敗させ、session cookieを削除する。ただし`admin_logout`は、信頼済みoriginからの要求に限り、匿名・期限切れ・失効済みsessionを冪等にcookie削除と匿名状態へ収束させる明示的な例外である。例外は業務状態を変更せず、sessionを復活させない。

## CSRF tokenの発行と復元

- OAuth callbackはCSRF tokenを生成し、照合用hashとServer保護鍵によるciphertextを管理sessionへ保存する。`303`ではCSRF tokenを返却しない。
- `admin_session_bootstrap`は有効sessionのciphertextをServer内で復号し、復号値のhashが同じsessionの照合用hashと定数時間比較で一致することを確認してから、平文を`200`応答で返す。平文をDB・cookie・ログ・URL・永続ブラウザ保存へ出さず、応答後にServerで保持しない。
- ciphertextの復号、hash照合、または保護記録の読取りに失敗した場合は`503 authentication_unavailable`とし、CSRF tokenを返さず、session cookie、アイドル期限、業務状態を変更しない。
- 同一sessionの有効期間中、bootstrapは同じCSRF tokenを返す。session失効時にhashとciphertextは認可に使えず、保守削除で回収する。

## 共通の安全性

- 管理認証操作は信頼済みアプリoriginだけに提供し、生成HTML隔離originから利用可能にしてはならない。
- Google Token、client secret、session署名鍵、DB接続文字列、認可code、`state`、`nonce`、PKCE verifier、cookie値、CSRF token、許可メールを、信頼済み管理UIのbootstrap応答以外のBrowser出力、生成HTML、公開表示、外部通信、分析イベント、エラー追跡、ログ、fixtureへ出さない。
- 保護記録を安全に利用できない場合はfail closedとし、管理操作を許可しない。
- Serverログは、操作種別、結果分類、相関ID、時刻だけを最低限記録する。Browserのエラーはアカウントの許可状態を区別しない一般化した表示にする。
- 認証状態またはCSRF tokenを返す応答、OAuth開始・callbackの成功・失敗応答は`Cache-Control: no-store`を必須とする。
- 管理認証操作は信頼済みアプリorigin以外へ`Access-Control-Allow-Origin`を返さず、資格情報を伴うcross-origin requestを許可しない。
- 生成HTML隔離originへの管理資格非流通の結合検証はFEAT-005の責務である。本Featureは信頼済みアプリが資格情報を非信頼表示境界へ渡さないことを境界テストする。
