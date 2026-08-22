# FEAT-002 テスト戦略

## 方針

Codex app-serverを実テストのoracleにしない。決定的なadapter test doubleで外部結果を制御し、domainからE2Eまで同じ受入れ条件を層ごとに検証する。実app-serverは資格情報を隔離した環境の最小smoke testだけで扱う。

## テスト層

| 層 | 対象 | 代表ケース |
| --- | --- | --- |
| domain / usecase unit | kind別入力、最大試行、失敗分類、候補不変条件 | 1回目成功、2回目成功、4回失敗、4回目後に呼ばない、revision元なし、`shutdown_rejected`をrequest-only T4へ一回だけ写像してretryしない。 |
| HTML validator unit | 完全HTMLの最小条件 | UTF-8空文字、doctype欠落、html/head/body欠落、parser不合格、合格文書。 |
| repository integration | PostgreSQL schema・transaction・constraint | migration順、attempt番号一意、状態整合CHECK、成功3 recordのatomicity、attempt付きT4とrequest-only T4の最終失敗atomicity、rollback。request-only T4は既存attempt履歴を変えず追加attemptなし。 |
| adapter contract | broker IPC・app-server v2 message・process境界 | private IPCの許可client・入力/出力制限、admit済みattemptごとの起動、`initialize`/`initialized`/`thread/start`/`turn/start`の固定順序・params、開始response前後のnotification保留・ID/順序照合、telemetry/errorの破棄/失敗化、単一agentMessageのHTML抽出、tool・複数/空出力・ID不一致・途中終了の`generation_unavailable`化、admission gateを直列化点とするclose先行（process/registry entryなしの`shutdown_rejected`）とadmit先行（一回の中断結果）、close/wait/terminate/kill/reap、内部ID非露出。DB attempt、T4、retry、HTTP status、DB書込み失敗は対象外。 |
| HTTP integration | request/responseとFEAT-001 guard | 201、400、409、422、404、401、403、503、`no-store`、候補本文不在、`running`の0〜3 attempt。 |
| UI component | 画面状態とA11y | required、busy、inline/form error、live region、focus、候補本文非表示。 |
| E2E | 管理者操作 | 初回成功、失敗後の新規生成、修正生成、再読込み、未認証。 |

## 受入れ条件トレーサビリティ

| 受入れ条件 | 主な証拠 |
| --- | --- |
| 自然言語topic・任意指示で初回生成 | HTTP/UI/E2Eでinitial requestとprompt port入力を確認。 |
| 修正元は合格済み版のみで、元版を変えない | usecaseのVersionSource fake、HTTP 409、integrationの書込みscope確認。 |
| HTML形式合格だけ候補化 | validator unitと成功transaction integration。 |
| 初回＋最大3再生成 | call-count unit、4 attempt integration、E2Eの最終失敗表示。 |
| 全試行履歴と安全な要約 | repository/HTTPで番号・分類・本文非露出を確認。 |
| 直後の新規要求 | UI/E2Eで422後に新しいrequest IDを作ることを確認。 |
| 認可・秘密隔離 | HTTP guard試験、adapter/HTTP/UIのresponse・log captureで秘密・HTML非露出を確認。 |
| 候補の隔離表示 | FEAT-005との結合E2E。FEAT-002単独の完了判定には含めない。 |

## 実行上の注意

PostgreSQL integrationは実migrationを適用した一時databaseで行い、mockだけでconstraint・transactionを代替しない。特にrequest-only T4が0〜3件の既存attempt履歴を許容し、新しいattempt行を追加せずrequestだけを原子的に終端化すること、T4書込み失敗時はrequestを`running`のまま残すことを確認する。adapter contract testは、private IPCの許可clientと入出力制限、shutdown時にbrokerのadmissionを閉じること、admissionとshutdownが競合したときclose先行ならprocess/registry entryなしの`shutdown_rejected`、admit先行ならprocess groupをreapして中断結果を一回だけ返すことだけを確認する。POST integrationはHTTP切断後にServer所有contextで実行を継続すること、前者をrequest-only T4、後者をattempt付きT4へ写像し、いずれもretryなしで422へ収束すること、T1〜T4書込み失敗時にretryせず500を返すことを確認する。E2Eはapp-server test doubleを使い、外部ネットワーク、実Codexアカウント、秘密値に依存させない。test fixtureには候補HTMLを必要最小限の無害な文書だけで置き、実ユーザー入力や資格情報を含めない。
