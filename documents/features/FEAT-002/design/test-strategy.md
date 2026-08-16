# FEAT-002 テスト戦略

## 方針

Codex app-serverを実テストのoracleにしない。決定的なadapter test doubleで外部結果を制御し、domainからE2Eまで同じ受入れ条件を層ごとに検証する。実app-serverは資格情報を隔離した環境の最小smoke testだけで扱う。

## テスト層

| 層 | 対象 | 代表ケース |
| --- | --- | --- |
| domain / usecase unit | kind別入力、最大試行、失敗分類、候補不変条件 | 1回目成功、2回目成功、4回失敗、4回目後に呼ばない、revision元なし。 |
| HTML validator unit | 完全HTMLの最小条件 | UTF-8空文字、doctype欠落、html/head/body欠落、parser不合格、合格文書。 |
| repository integration | PostgreSQL schema・transaction・constraint | migration順、attempt番号一意、状態整合CHECK、成功3 recordのatomicity、最終失敗atomicity、rollback。 |
| adapter contract | app-server v2 message・process境界 | attemptごとの起動、`initialize`/`initialized`/`thread/start`/`turn/start`の固定順序・params、開始response前後のnotification保留・ID/順序照合、telemetry/errorの破棄/失敗化、単一agentMessageのHTML抽出、tool・複数/空出力・ID不一致・途中終了の`generation_unavailable`化、close/wait/terminate/kill、内部ID非露出。 |
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

PostgreSQL integrationは実migrationを適用した一時databaseで行い、mockだけでconstraint・transactionを代替しない。adapter contract testは、HTTP切断後もServer所有contextで実行を継続することと、shutdown時に新規attemptを起動せず子processをreapすることを確認する。E2Eはapp-server test doubleを使い、外部ネットワーク、実Codexアカウント、秘密値に依存させない。test fixtureには候補HTMLを必要最小限の無害な文書だけで置き、実ユーザー入力や資格情報を含めない。
