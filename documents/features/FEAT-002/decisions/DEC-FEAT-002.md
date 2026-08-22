# DEC-FEAT-002 — 生成要求の公開・共有・運用契約

状態: **承認済み（L3、2026-08-16）**

## 判断が必要な理由

FEAT-002は管理HTTP、候補と合格済み版の参照、HTML合格条件、Codex execution brokerの運用設定を追加する。これらは要件から目的は導けるが、wire形式・共有ID・運用上の既定を一意には導けない公開または共有契約である。

## 推奨案 A

次を一体の契約として採用する。

1. 管理APIは同期の`POST /admin/generation-requests`と再読込み用の`GET /admin/generation-requests/{id}`を提供する。POSTは形式合格なら201、最大4試行の全失敗なら422と安全な結果記録を返す。認可はFEAT-001のsession / Origin / CSRF契約を使う。
2. 生成要求、試行、候補の識別子はUUIDとする。修正要求は`source_version_id` UUIDでFEAT-003の参照可能な合格済み版を指定し、Server内の`VersionSource`だけが元HTMLを取得する。候補IDは管理APIの不透明metadataだが、画面へ文字列として表示しない。
3. 合格HTMLは、空でないUTF-8文字列をHTML5 parserで解析でき、明示的な`<!doctype html>`と開始・終了する`html`、`head`、`body`要素を持つ完全HTML文書とする。内容の品質や外部URLは検査しない。
4. Go Serverは必須の`TOPIC2HTML_CODEX_EXECUTION_BROKER_ENDPOINT`を起動時に検証し、brokerだけが必須の`TOPIC2HTML_CODEX_APP_SERVER_EXECUTABLE`と`TOPIC2HTML_CODEX_APP_SERVER_WORKDIR`を検証して固定argv `app-server --stdio`で起動する。brokerはGo Serverと異なる専用service accountでCodex資格情報を扱い、両者はprivate local IPCだけで接続する。この配置は[DEC-FEAT-003](DEC-FEAT-003.md)で承認済みである。

## 代替案 B

生成を非同期job APIにし、POSTは受付IDだけを返し、別途pollingまたは通知で完了を取得する。HTML形式はparserによる受理だけに緩め、candidate/versionのID形式とapp-server起動方法は後続Featureまたは実装で決める。

## 影響と推奨理由

| 選択      | 利用者・実装への影響                                                                                                                                   |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| A（推奨） | 管理者は送信画面で完了または安全な失敗を直接確認できる。FEAT-003/005はUUID候補参照を前提に設計でき、検証済み候補の意味と秘密隔離が実装前に固定される。 |
| B         | 長時間処理に強いjob基盤を追加できるが、状態照会・通知・取消・job保持の要件と契約が新たに必要になる。合格判定・ID・運用は後続設計へ持ち越される。       |

要件は完了時間を保証せず、job基盤・通知・取消を求めていない。既存のFEAT-001管理HTTP契約を最小限に拡張し、FEAT-003/005との責務境界を早期に固定できるためAを推奨する。

## 承認後の影響

利用者は選択肢Aを承認した。詳細設計のHTTP、DB、operation、画面、設定資料はこの契約で確定し、設計レビューを再実施する。参考として、Bを採る場合は非同期jobの状態、取得operation、保持と取消、画面遷移を改めて設計し、FEAT-003/005の共有契約も再調整する。
