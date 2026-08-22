# FEAT-002 詳細設計レビュー

実施日: 2026-08-16  
判定: **pass**

## 独立監査の対象と正規ソース

既存のreview・readiness reviewの結論は根拠にせず、R-001補正後の詳細設計と限定再同期済みTask/handoffを、以下の正規ソースへ独立に照合した。

| 区分 | 正規ソース | 確定状況 | 適用範囲 |
| --- | --- | --- | --- |
| 機能要件 | `requirements/requirements.md` REQ-001〜014 | confirmed | Codex生成、HTML検証、最大再生成、試行記録、安全な失敗表示。 |
| 業務規則・制約 | `requirements/business-rules.md` BR-002〜005・016、`requirements/nonfunctional-requirements.md` NFR-001・003〜006、`requirements/constraints.md` CON-001・003・004・ASM-001 | confirmed / L2 assumption | 失敗の統一処理、初回を含む最大4試行、秘密非露出、正常生成の完了時間を保証しない。 |
| 初期設計 | `design/architecture.md`、DEC-ARCH-001〜003 | approved | Server側外部I/O、設定注入、repository境界。 |
| 先行Feature契約 | `features/FEAT-001/design/http-contract.md`、`features/FEAT-001/implementation-handoff.yaml` | approved / ready | 管理read/mutation guard、session idle期限更新、migration・composition基盤。 |
| Feature Decision | `features/FEAT-002/decisions/DEC-FEAT-001.md`、`DEC-FEAT-002.md` | approved (L3) | 候補所有、同期HTTP、UUID、HTML合格条件。 |
| 実行隔離Decision | `features/FEAT-002/decisions/DEC-FEAT-003.md` | approved (L3) | 別OS service accountのprivate local IPC broker、双方向の秘密分離。 |
| shutdown Decision | `features/FEAT-002/decisions/DEC-FEAT-004.md` | approved (L3) | brokerのadmission/registry/process group、`shutdown_rejected`とadmit済み中断の区別、retryなし終端。 |

未承認のL3/L4 Decisionは検出しなかった。IPC transport、内部package・symbol・test helperは、承認済みのprivate local IPC、OS権限、入出力制限を満たす限りImplementation領域の選択である。

## 契約根拠トレーサビリティ

| 契約領域 | 設計上の契約 | 根拠の判定 | 監査結果 |
| --- | --- | --- | --- |
| 管理HTTP | 同期POST/GET、UUID、201/422、FEAT-001 guard、候補HTML非露出 | explicit / derived | `http-contract.md`、operation、画面、DB資料で整合する。 |
| 永続化・transaction | request・admit済みattempt・candidate、T1〜T4、候補不変 | derived | REQ/BR、DEC-ARCH-003、FEAT-001と整合する。外部I/O中にDB transactionを保持しない。 |
| request-only T4 | `shutdown_rejected`では新しいattempt行を作らず、既存の0〜3件のT2履歴を変えず、requestだけを終端化する | explicit | DEC-FEAT-004 L2補正、`database-schema.md`、operation、POST契約、TASK-002-01/03、test strategyで一致する。 |
| attempt付きT4 | 4回目またはadmit済みattemptのshutdown中断ではfailed attemptとrequestを同一transactionで終端化する | explicit | DEC-FEAT-004、DB access map、operation、TASK-002-01/03、POST契約で一致する。 |
| admission / retry / shutdown | brokerだけが直列化点を持つ。close先行は`shutdown_rejected`、admit先行は一回だけの中断結果であり、POSTだけがT4形態を選び、いずれもretryしない | explicit | DEC-FEAT-004、adapter契約、operationの責務表、tasks/handoffで一致する。 |
| 実行・認証分離 | Go Serverはprivate IPC client、brokerが別OS accountでCodex認証、固定argv、workdir、process groupを所有する | explicit | DEC-FEAT-003、runtime、adapter契約、Feature設計で整合する。 |
| 公開・秘密 | 同一origin JSON、`no-store`、候補HTML・prompt・内部ID・外部詳細・秘密値を非露出 | explicit / derived | HTTP、runtime、adapter、画面、test strategyで整合する。 |
| TASK-002-02の責務 | private IPC、admission、registry、process cleanup、HTMLまたは安全な結果一回だけ | explicit | adapter契約、TASK-002-02、handoff、adapter contract testがDB/T4/retry/HTTPを明示的に対象外とする。 |
| TASK-002-01/03の責務 | TASK-002-01は両T4とrollbackをtransactionで検証し、TASK-002-03はbroker結果のPOST写像、retry停止、422/500、HTTP切断後継続を検証する | explicit | operationの責務表、tasks/handoff、test strategyで一致する。 |

## 資料間整合の監査

- `design.md`、adapter、runtime、HTTP、operation、DB、test strategyは、Go Serverがprivate IPC endpointだけを検証し、brokerが実行可能ファイル・workdir・Codex資格情報・子processを所有する点で整合する。Go Serverがapp-serverを直接起動する記述は検出しなかった。
- `shutdown_rejected`とadmit済みshutdown中断の永続化契約は、request-only T4とattempt付きT4を明確に分け、公開HTTPではいずれもretryなしの422 `generation_failed`へ収束する。公開I/O、DB field、状態遷移、rollback、秘密非露出に未解決の矛盾は検出しなかった。
- R-001補正後、adapter contract/testのclose先行ケースはprocess未起動・registry未登録と`shutdown_rejected`だけを観測する。DB attempt/T4、retry、HTTP status、DB書込み失敗、HTTP切断後の継続は対象外であり、TASK-002-01/03とtest strategyへ移管済みである。brokerにDB観測またはT4選択を委ねる記述は検出しなかった。
- 管理画面の到達・退出、loading／success／failure／`running`、POST/GET、候補HTMLと秘密の非露出は、要件からFeature設計、画面、operation、HTTPへ追跡できる。隔離表示はFEAT-005の共同受入れ条件であり、当Featureの範囲外として一貫している。
- 同期済みの`tasks.md`と`implementation-handoff.yaml`は、現行Decision・詳細設計・review/承認状態を`stale_pending_re_review`として正しく参照し、未承認契約を追加していない。

## 反証的確認

- shutdown先行を通常の`generation_unavailable`としてretryする案は、停止後の新規process起動を許すためDEC-FEAT-004に反する。設計の`shutdown_rejected`終端は妥当である。
- brokerがfailed attemptを合成する案は、request-only T4が外部interactionなし・新しいattempt行なしである契約と、DB/T4がGenerationStoreとPOSTの責務である境界に反する。設計はこれを許可していない。
- adapter contract testがDB履歴またはT4書込み失敗を観測する案は、brokerの所有範囲を越える。現行資料はこれらをTASK-002-01のPostgreSQL integrationとTASK-002-03のusecase/POST integrationへ配置しており、矛盾を解消している。
- cleanupの5秒はshutdown時の猶予であり、正常generationのdeadlineではない。NFR-006との矛盾はない。

## 総合判定と次のゲート

**pass**。R-001補正後、公開HTTP、request-only/attempt付きT4のDB契約、retry停止、秘密隔離、brokerのadmission/cleanup、Task/handoffの責務境界は、正規要件と承認済みDecisionに整合する。実装者またはTaskへ仕様再設計を委ねる未解決事項は検出しなかった。

次のゲートは **implementation-readiness-review**。handoffを実装可能にするには、その独立監査のpassと利用者による詳細設計の明示承認、通常のDesign Readiness Audit・handoff最終化が必要である。
