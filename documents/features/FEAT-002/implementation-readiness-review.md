# FEAT-002 実装開始可能性レビュー

実施日: 2026-08-16  
判定: **pass**

## 独立監査の範囲

設計者および詳細設計レビュー担当者とは独立に、`requirements.md`、`design.md`、`design/**`、承認済みのDEC-FEAT-001/002、`design-review.md`、要件・Initial Design・FEAT-001の管理HTTP契約、workflow stateとDecision Policyを実装者視点で照合した。

FEAT-002のTaskとimplementation handoffはまだ存在しない。これは本ゲートの後段であり、本レビューの欠落ではない。

## 実装者が依存する確定契約

| 領域 | 監査結果 | 実装可能な契約 |
| --- | --- | --- |
| PostgreSQL / migration | pass | migration `003`の順序・transaction失敗時rollback、3 tableのDDL相当、UUID・NULL・CHECK・FK・index、候補不変trigger、T1〜T4が定義されている。 |
| 管理HTTP | pass | POST/GETのmethod・path、JSON入力、正規化、201/400/409/422/404/500、resource/error field、`no-store`、FEAT-001のread/mutation guardが定義されている。 |
| 画面 | pass | 初回・修正文脈、結果再表示route、送信中・成功・最終失敗・中断記録・session障害、HTTP写像、A11y、responsive、候補本文/秘密の非露出が定義されている。 |
| operation / transaction | pass | 最大4試行、冪等性の扱い、VersionSource失敗、DB rollback時の停止、外部I/O中にtransactionを保持しない規則、`running`の安全な観測が定義されている。 |
| 設定・秘密値 | pass | Server限定の実行可能ファイルと専用workdir、起動時fail-fast、固定argv、資格情報の所有境界、Browser・DB・ログへの非露出、smoke testの責務が定義されている。 |
| Codex app-server外部境界 | pass | attemptごとの子process所有、v2 JSON-RPCの固定wire・相関・許可notification、出力採用規則、失敗正規化、HTTP切断、shutdown、close/wait/terminate/kill/reapの順序が定義されている。 |
| 検証 | pass | unit / PostgreSQL integration / adapter contract / HTTP / UI / E2E / smoke testの役割、test doubleの観測点、fixtureと秘密値の制約が定義されている。 |

## 外部adapterの実装可能性

以前の差し戻し事項だったadapter境界は、`design/codex-app-server-adapter.md`で次のように具体化されている。

- 一attempt一processであり、同時要求間でstdio、thread、turnを共有しない。
- `initialize`、`initialized`、`thread/start`、`turn/start`の送信順、固定params、request IDとthread/turn IDの相関、開始responseより先に来たnotificationの保留・再照合が定義されている。
- 空の専用workdir、`approvalPolicy: never`、`sandbox: read-only`を固定し、client request、approval、command/file/MCP等の非許可item、未知または順序不正wireを`generation_unavailable`へfail closedする。
- `turn.status = completed`と、一件だけの`item/completed.agentMessage.text`を両方必要とする。delta、reasoning、tool結果、複数/空出力、途中終了は候補に採用しない。
- 通常処理にdeadlineを置かず、HTTP切断後はServer所有contextで処理を継続する。終了時のstdin close、wait、process groupへのterminate/kill、reap、新規attempt停止とshutdown中の失敗確定が定義されている。

これにより、process再利用、wireの許容範囲、HTML本文の選択、cleanupをrepository実装時の裁量に残していない。5秒は正常処理の期限ではなくcleanup猶予として限定され、NFR-006とも整合する。

## 実装への不適切な委譲の確認

公開HTTP、候補/版の所有境界、HTML合格条件、DB状態、retry、設定、認可、外部wire、安全な失敗・cleanupに関する「実装時に決める」記述は検出しなかった。残るfile path、package、symbol、JSON-RPC reader/writerの局所構造はImplementation領域で決めてよい範囲である。

FEAT-003のVersionSource実装とFEAT-005の隔離表示は後続Featureの明示的責務であり、FEAT-002はport/fakeおよび候補metadataで独立して実装・検証できる。これは未決定の依存ではない。

## 総合判定と次のゲート

**pass**。追加の製品・公開・security判断なしに、FEAT-002のDB、HTTP、画面、生成operation、Codex app-server adapter、設定、テストを実装開始できる。

次のゲートは利用者による詳細設計レビューである。承認後に`task-breakdown`へ進む。
