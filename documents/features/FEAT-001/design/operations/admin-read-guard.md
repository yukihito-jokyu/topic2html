# `admin_read_guard` — 管理読取りの認可ガード

## 役割

後続Featureの管理プレビュー、履歴、任意版確認など、公開閲覧ではない全read operationの直前に適用する。公開済みHTMLの匿名閲覧には適用しない。

## 契約

有効sessionは存在、未失効、絶対・idle期限内、認可済みメールと現行許可メールの完全一致を満たす。満たせば後続read operationへ通過する。満たさなければHTTP契約に従う`401 unauthenticated`とし、業務データを返さず、期限切れ・失効・不一致のcookieは削除する。保護記録を読めない場合もfail closedで`503 authentication_unavailable`とする。

このガードは読取りのため`Origin`とCSRF tokenを要求せず、sessionのidle期限を延長しない。生成HTML隔離originの表示実現はFEAT-005の責務だが、管理読取りにこのガードを適用する契約はFEAT-001が所有する。
