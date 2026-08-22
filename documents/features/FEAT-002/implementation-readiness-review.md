# FEAT-002 実装開始可能性レビュー

実施日: 2026-08-16  
判定: **pass**

## 独立監査の範囲

設計者、R-001補正後の詳細設計レビュー担当者、Task/handoff同期担当者とは異なる実装者視点で、`requirements.md`、詳細設計、承認済みDEC-FEAT-001〜004、`design-review.md`、`tasks.md`、`implementation-handoff.yaml`、Initial Design、FEAT-001のhandoffおよび現行Go Server compositionを照合した。

監査の焦点は、broker/private IPCの設定・秘密所有、admission/shutdown/retryの排他、process group cleanup、test double、既存compositionへの接続である。実装の内部package、IPCの具体実装方式、test helperは、承認済みの安全契約を変えないImplementation領域の選択として扱った。

## 実装者が依存する確定契約

| 領域 | 判定 | 実装可能な契約 |
| --- | --- | --- |
| broker / private IPC | pass | Go Serverはprivate local IPC endpointだけを起動時に検証してbroker clientへ注入する。brokerだけが別OS service accountでCodex認証、実行可能ファイル、空workdir、app-server子processを所有する。clientは完成済みpromptだけを送り、HTML本文または安全な分類だけを受け取る。任意command・argv・cwd・環境・認証path・外部ID/詳細errorをIPC入出力に含めない。 |
| 設定・秘密 | pass | `TOPIC2HTML_CODEX_EXECUTION_BROKER_ENDPOINT`はGo Server所有、実行可能ファイルとworkdirはbroker所有である。双方で欠落・形式不正・接続/実行不能をfail-fastとし、Go ServerのDB/Google/CSRF/session秘密とCodex認証の双方向非継承を明示している。Browser、DB、ログ、fixtureへの非露出も定義済みである。 |
| app-server wire | pass | broker内で固定argv、`initialize`→`initialized`→`thread/start`→`turn/start`、固定params、ID照合、保留notificationの再照合、許可notification、単一完了agentMessageの採用規則、未知/不正wireのfail closedが定義済みである。 |
| admission / retry / shutdown | pass | brokerの唯一の直列化点が`open → closing → closed`を所有する。close先行は`shutdown_rejected`、process・attempt行なしでrequestだけをT4終端化する。admit先行はregistry登録済みattemptを一回だけ中断結果として返し、T4でattemptとrequestを終端化する。いずれも後続retryを開始しない。Go Serverはgate・registry・processを更新しない。 |
| cleanup / 永続化 | pass | 各attemptは独立process groupであり、stdin close・送信停止、5秒wait、group terminate、5秒wait、group kill、wait/reapの順を一回だけ実行する。brokerのreapはT4書込み失敗から独立し、書込み失敗時は既存規則どおり`running`を残して再開しない。外部I/O中にDB transactionを保持しない。 |
| テスト | pass | 決定的doubleがIPC許可client・入出力制限、固定wire、順序/ID異常、候補選択、切断後継続、close先行/admit先行の競合、group cleanup、終端記録失敗後のreapを観測する。通常CI/E2Eは実Codex認証・外部networkへ依存せず、実環境だけで秘密隔離smoke testを行う。 |
| Server composition | pass | 現行の`backend/cmd/server`は設定読取り・検証、repository/usecase/handlerのconstructor注入、HTTP Server生成をcompositionに限定している。FEAT-002はここへbroker endpoint検証、broker client注入、HTTP受理停止後にbroker cleanupとT4処理を待ってからDB依存を閉じる停止順を追加できる。Go Serverからbroker実行可能ファイル・workdir・Codex認証を読ませる変更は契約違反である。 |

## 反証的確認

- Serverがbrokerを直接子processとして起動する案は、既存のGo Server環境・UIDの継承によりDEC-FEAT-003の秘密分離に反するため採用できない。private IPC brokerという承認済み境界がこの反証を排除している。
- Go Serverとbrokerの両方にadmission gateを置く案は、T2後のretryとshutdownの順序を二重に管理し、attempt/終端記録の重複を生む。設計はbrokerだけを直列化点とし、この競合を一意にしている。
- shutdownを通常の`generation_unavailable` retryへ戻す案は、停止後の新規process起動を許し得る。`shutdown_rejected`とadmit済み中断を分け、どちらもretryなしのT4とするため、この抜け道はない。
- 親PIDだけを終了する案は孫processを残し得る。group単位のsignalと最終reapの順序、ならびにそれを観測するdoubleが定義されている。
- 現行Serverにはまだgraceful shutdown実装がないが、これは実装対象であって未決定の製品・運用契約ではない。停止順、待機対象、DB依存のclose条件はDEC-FEAT-004とadapter契約に固定済みである。

## Task / handoff 整合

`TASK-002-02`とhandoffはDEC-FEAT-003/004を参照し、旧来の「Go Serverがapp-serverを直接起動する」契約を含まない。Taskはbroker client、owner別fail-fast設定、OS identity、private IPC、admission、process group cleanup、決定的doubleを明示し、R-001補正どおりDB attempt/T4、retry、DB書込み失敗、HTTP client切断後の継続、HTTP statusを対象外として分離している。これらはTASK-002-01/03のPostgreSQL・usecase・POST integration検証に明示的に配置されている。

workflow stateの`stale`は、承認済みDecision後に再レビュー・利用者承認・handoff最終化を要求する進行状態であり、契約内容の未承認やTask/handoffの不整合を意味しない。未解決のL3/L4 Decisionは検出しなかった。

## 実装への不適切な委譲の確認

公開HTTP、DB、brokerの入力/出力制限、秘密所有、admissionとshutdownの順序、retry排他、cleanup、test doubleの観測点に「実装時に決める」べき未確定契約は検出しなかった。private IPCの具体transport、broker binaryの内部配置、同期・mutex・process APIなどは、same-host private endpoint・OS permission・上記の入出力制限を満たす限り局所的な実装選択である。

## 総合判定と次のゲート

**pass**。追加の製品、公開、security、運用Decisionを行わずに、FEAT-002を実装できる。特にTASK-002-02は、現行Server compositionを正規の設定/注入/停止順で拡張し、brokerを別OS identityの外部境界として実装できる。

次のゲートは、更新後の詳細設計に対する利用者承認である。承認後にTask Breakdownおよびhandoff最終化を行う。
