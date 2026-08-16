# FEAT-002 詳細設計レビュー

実施日: 2026-08-16  
判定: **pass**

## 独立監査の対象と正規ソース

既存レビューの結論には依存せず、Codex app-server adapterを含むFEAT-002詳細設計を次の正規ソースと照合した。

| 区分 | 正規ソース | 確定状況 | 適用範囲 |
| --- | --- | --- | --- |
| 機能要件 | `requirements/requirements.md` REQ-001〜014 | confirmed | Codex生成、HTML形式検証、再生成、試行記録、安全な失敗表示。 |
| 業務規則・制約 | `requirements/business-rules.md` BR-002〜005・016、`requirements/constraints.md` CON-001・003・004・ASM-001 | confirmed / assumption | 取得不能と形式不合格の統一、初回を含む最大4試行、秘密を含まない要約。 |
| 初期設計 | `design/architecture.md`、`design/cross-cutting-concerns.md`、DEC-ARCH-001〜003 | approved | repositoryに閉じる外部I/O、`cmd`からの設定注入、信頼境界。 |
| 先行Feature契約 | `features/FEAT-001/design/http-contract.md` の管理read/mutation guard | approved | 管理HTTP guardとmutation時のsession idle期限更新。 |
| Feature Decision | `features/FEAT-002/decisions/DEC-FEAT-001.md`、`DEC-FEAT-002.md` | approved（L3） | 候補の所有、同期HTTP、UUID、HTML合格条件、固定argvとServer限定workdir。 |
| app-server wire | `codex app-server generate-json-schema`で取得したv2 schema | implementation-facing evidence | `initialize`、`thread/start`、`turn/start`、responseとnotificationのfield形。 |

## 契約根拠トレーサビリティ

| 契約領域 | 設計上の契約 | 根拠の判定 | 監査結果 |
| --- | --- | --- | --- |
| 管理HTTP・画面 | 同期POST/GET、UUID、201/422、FEAT-001 guard、候補HTML非露出 | explicit / derived | `http-contract.md`、operation、画面、DB資料は整合する。 |
| 永続化・transaction | request・attempt・candidate、最大4試行、T1〜T4、候補不変 | derived | REQ、BR、DEC-ARCH-003、FEAT-001と整合する。外部I/O中にDB transactionを保持しない。 |
| process所有・設定 | attemptごとの独立process、固定argv、専用workdir、認証値非露出 | explicit / derived | DEC-FEAT-002、Initial Designと整合する。並行要求のstdio共有は発生しない。 |
| v2開始wire | `initialize`→`initialized`→`thread/start`→`turn/start`、最小params、response ID検査 | derived | schemaの必須fieldと一致する。`thread/start`で得たthread IDを保持し、`turn/start` responseには存在しないthread IDを読まない。 |
| 相関・notification | 保留後にthread/turn IDを再照合し、telemetryは破棄、異常wireは安全に失敗化 | derived | `TurnStartResponse`の`turn.id`と送信済みthread IDを保持し、以後のscoped notificationの`threadId`とturn IDを照合するため、response/notificationのinterleaveを実装者へ委譲していない。 |
| 候補選択 | completed済みの単一`agentMessage.text`だけをvalidatorへ渡し、delta・tool・reasoningを採用しない | derived | v2の`AgentMessageThreadItem`の単一`text`とREQ-005、CON-003に整合する。複数・空・非textを安全に失敗化する規則も明確である。 |
| 失敗・終了 | deadlineなし、HTTP切断後もServer所有contextで継続、close→wait→terminate→kill→reap | derived / L2 | NFR-006と矛盾しない。5秒は正常生成の期限でなくcleanup猶予であり、子process回収とshutdown中の新規起動停止が明記される。 |

## 資料間整合の監査

- 画面のPOST結果画面への遷移とreload GET、HTTPの`running`表現、operationのT1〜T4、DBのstate／attempt制約は一致する。候補HTMLはHTTP・DOM・安全ログに出さず、FEAT-003の採用とFEAT-005の隔離表示へ越境していない。
- adapterはattemptごとのprocess所有、固定wire、approval／tool itemの拒否、開始response前のnotification保留と再照合、`thread/status/changed`・`thread/settings/updated`・`thread/tokenUsage/updated`・`error`の扱い、出力選択、cleanupを定義している。
- 特に`TurnStartResponse`は`turn`だけを返しthread IDを返さないv2 schemaと整合する。送信済み`turn/start.params.threadId`を保持threadとして使い、responseではrequest IDと`turn.id`だけを受理し、`turn/started`、item、token usage、error、`turn/completed`のnotificationで両IDを照合する。responseとnotificationの到着順による候補混入はない。
- `turn/completed`の`completed`と、一件だけの`item/completed.agentMessage.text`を両方必須にするため、delta、turn response内のitems、reasoning、tool結果をHTML候補として誤採用しない。
- processの終了順とtest doubleの観測点はadapter資料・テスト戦略・実行時設定で一致する。子process groupをreapするまでをadapter所有とし、外部ID・エラー本文・telemetryを永続化、HTTP、安全ログから除外する。

## 人間Decisionの境界

公開API、UUID、候補／版の責務、HTML合格条件、固定argvとworkdirはDEC-FEAT-001/002で承認済みである。adapterのresponse相関、notification処理、候補抽出、cleanup猶予は、承認済み外部境界を変更せず既存v2 schemaとNFR-006から具体化できるL1/L2詳細である。未承認のL3/L4判断は検出しなかった。

## 総合判定と次のゲート

**pass**。公開・永続化・設定・外部wireの契約根拠、資料間整合、失敗時の安全な扱いを確認した。次のゲートは`implementation-readiness-review`であり、Taskまたはimplementation handoffの作成はまだ行わない。
