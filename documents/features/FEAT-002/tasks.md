# FEAT-002 実装タスク — 解説HTMLの生成と失敗復旧

## Task一覧

| ID | タイトル | 論理領域 | 依存 |
| --- | --- | --- | --- |
| TASK-002-01 | 生成要求のmigration・永続化基盤を実装・検証する | migration・生成記録永続化 | FEAT-001の実装完了 |
| TASK-002-02 | Codex app-server adapter基盤を実装・検証する | 外部生成境界 | FEAT-001のServer設定・composition基盤 |
| TASK-002-03 | `POST /admin/generation-requests` をend-to-endで実装・検証する | 生成制御・管理mutation・画面送信 | TASK-002-01、TASK-002-02、FEAT-001の管理mutation guard |
| TASK-002-04 | `GET /admin/generation-requests/{id}` をend-to-endで実装・検証する | 要求記録読取り・管理read・画面再表示 | TASK-002-01、TASK-002-03、FEAT-001の管理read guard |

## 分割方針

`POST /admin/generation-requests` と `GET /admin/generation-requests/{id}` は、それぞれ入力から出力、画面、HTTP検証までを一つのAPI Taskに完結させる。handler、usecase、repository、画面、HTTP testなどの技術層で同一APIを分割しない。

一方、migration・生成記録永続化とCodex app-server adapterは、個別APIの入力・出力を実装しない再利用可能な横断基盤として独立Taskに戻す。前者はPOST/GETの共通データ契約、後者は将来の生成operationにも再利用する外部境界を提供する。POST Taskは両基盤へ依存し、POST固有の生成制御・HTML検証・入力/出力・画面/HTTP検証を一体で実装する。GET Taskは永続化基盤とPOSTの結果導線へ依存し、GET固有の読取り・入力/出力・画面/HTTP検証を一体で実装する。

## TASK-002-01 生成要求のmigration・永続化基盤を実装・検証する

**目的:** 生成要求、試行履歴、形式合格済みHTML候補を、安全なtransaction・rollback・不変性を持つ共通の永続化契約として提供する。

**関連要件:** REQ-005、REQ-007〜014、BR-003〜005、BR-016、CON-003、CON-004

**作業内容:** migration `003`、生成要求・attempt・candidateの永続化契約、T1〜T4 transaction、読取りaccess map、候補不変性を[DB schema](design/database-schema.md)とoperation資料に従って実現する。POST/GETのHTTP handler、画面、Codex呼出し、生成制御は含めない。

**受け入れ基準:**

- migration `003`、`generation_requests`、`generation_attempts`、`generated_html_candidates`、制約、index、候補不変triggerが[DB schema](design/database-schema.md)に一致する。migrationはtransactionで適用され、失敗時に全変更をrollbackする。
- T1は管理sessionのidle期限更新と`running` request、T2は1〜3回目の失敗attempt、T3は成功attempt・candidate・成功request、T4は4回目失敗attempt・最終失敗requestを、設計どおりのtransactionで確定する。外部I/O中にDB transactionを保持しない。
- T1後の書込み失敗では後続状態を作らず、T1失敗時は外部呼出しを開始できない。T2〜T4の書込み失敗時は候補または後続attemptを追加せず、rollback規則を守る。
- `(generation_request_id, attempt_number)`の一意性、attempt番号1〜4、request・attemptの状態整合、candidateの1対0..1、候補HTML非空、候補のUPDATE/DELETE拒否、request/attempt/candidateのread-only取得が設計どおりに検証される。
- 実PostgreSQL integrationでmigration、制約、index、不変trigger、T1〜T4、rollback、read access mapを検証する。候補HTML、未検証出力、Codex内部ID・詳細errorをGET用metadata以外の読取り結果へ出さない。

**依存:** FEAT-001が提供するPostgreSQL migration runner、管理session、Server compositionと設定注入基盤。

**対象外・注記:** POST/GETの入力検証・HTTP応答・画面、Codex app-server呼出し、HTML妥当性判定、生成再試行のorchestrationは対象外である。合格候補の採用、version/content/公開状態、VersionSourceの実装はFEAT-003が所有する。

## TASK-002-02 Codex app-server adapter基盤を実装・検証する

**目的:** 一試行ごとに独立したCodex app-server子processを安全に管理し、検証前の生成本文または安全な生成不能分類だけを上位の生成operationへ返せる外部境界を提供する。

**関連要件:** REQ-004、REQ-009、REQ-010、BR-002、BR-016、NFR-003〜006、CON-001、CON-004

**作業内容:** Server限定の起動時fail-fast設定、固定argv・専用workdir、固定JSON-RPC wire、通知状態機械、候補本文選択、切断後継続、shutdown cleanupを[adapter契約](design/codex-app-server-adapter.md)に従って実現する。POST/GETのHTTP handler、画面、生成要求の永続化、HTML妥当性判定は含めない。

**受け入れ基準:**

- 実行可能ファイルとworkdirはServer限定設定として起動時fail-fast検証され、各attemptは固定argv `app-server --stdio`、専用空workdir、`approvalPolicy: never`、`sandbox: read-only`で独立processとして起動する。利用者入力・環境由来の値をargv、cwd、環境変数へ追加しない。
- `initialize`、`initialized`、`thread/start`、`turn/start`の固定wire、response ID・thread/turn/item IDの照合、開始response前の通知保留と再照合、許可された通知だけを受理する状態機械がadapter契約に一致する。
- 完了済み単一agentMessageの空でない本文だけを未検証本文として返す。複数・空・非text agentMessage、tool/file/MCP等の非許可item、wire/ID/順序異常、stdio異常、非retry error、turn失敗は`generation_unavailable`として安全に分類する。
- 外部ID・詳細error・資格情報・未検証本文をログ、永続化、HTTP、画面へ出さない。adapterはapp-server認証値を読み取らず、専用service accountの実行環境だけが扱う。
- HTTP client切断後もServer所有contextでattemptを継続でき、shutdown時は新規attemptを開始せず、stdin close、wait、terminate、kill、reapの順に子processをcleanupする。正常generationに完了deadlineを置かない。
- 決定的test doubleによるadapter contract testで、固定wire、通知順序、候補選択、異常分類、切断後継続、shutdown cleanupを検証する。通常CIは実Codex認証情報または外部ネットワークに依存しない。

**依存:** FEAT-001が提供するServer設定・composition基盤。

**対象外・注記:** request/attempt/candidateの永続化、HTML parserによる形式判定、最大4回の再試行、POST/GETのHTTP応答・画面は対象外である。

## TASK-002-03 `POST /admin/generation-requests` をend-to-endで実装・検証する

**目的:** 管理者が初回または修正の生成要求を同期的に開始し、最初のHTML形式合格結果を安全な候補として確定するか、最大4試行後の安全な失敗として終了できるようにする。

**関連要件:** REQ-001〜005、REQ-007〜014、REQ-026、REQ-027、BR-002〜005、BR-016、CON-001、CON-003、CON-004、ASM-001

**作業内容:** `POST /admin/generation-requests` を、承認済みのmutation guard、kind別入力規則、VersionSource照合、生成制御、HTML形式検証、共通永続化・Codex adapterの利用、管理画面の送信状態まで一貫して実現する。operation資料のT1〜T4、HTTP契約、画面設計、runtime configurationに従い、POSTの成功・失敗を同じTask内で自動検証する。

**受け入れ基準:**

- POSTはmutation guard通過後にkind別JSONを正規化・検証する。initialはtrim後非空topicと任意instructions、revisionは参照可能な合格済みversionとtrim後非空instructionsだけを受け付け、修正元が利用不可ならrequest作成・外部生成を行わず409 `source_version_not_available`を返す。
- 共通基盤を用いて、空でないUTF-8文字列をHTML5 parserで解析でき、明示的doctypeと開始・終了するhtml/head/bodyを持つ完全HTML文書だけを合格とする。取得不能と形式不合格は同じ再試行フローで順に記録し、初回を含め最大4回で停止する。最初の合格時だけ候補を保存し、全失敗時は候補を作らず安全な定型要約で終了する。
- POSTは成功時201、入力不正400、修正元利用不可409、全試行失敗422、内部障害500を[HTTP契約](design/http-contract.md)どおりに返す。各POST requestは別の生成要求として記録し、最終応答に`running`を返さない。JSON、`Cache-Control: no-store`、FEAT-001の401/403/503 error envelopeとsession副作用規則を再利用する。
- 管理画面で初回の必須topic・任意instructions、修正の非編集summary・必須instructions、送信中の二重送信防止、201成功、400/409/422/500/401/403/503の安全な表示、422後の新規要求を[画面設計](design/screen-specification.md)どおりに提供する。候補HTML、candidate ID、source version ID、資格情報、session/CSRF値、Codex詳細をDOM、URL、console、analytics、永続ブラウザ保存へ出さない。
- unit、POST HTTP integration、画面component/E2Eで、初回成功、修正成功、修正元利用不可、HTML不合格・取得不能の最大4試行、認証/Origin/CSRF拒否、busy、候補本文非露出を[テスト戦略](design/test-strategy.md)どおりに検証する。通常CI/E2Eは実Codex認証情報や外部ネットワークに依存しない。

**依存:** TASK-002-01、TASK-002-02、FEAT-001の管理mutation guard。

**対象外・注記:** 合格候補の採用、version/content/公開状態の変更、VersionSourceの実装は作らない。候補の実表示・隔離origin・取消・非同期job・polling/通知も対象外である。VersionSourceはFEAT-003が後続で実装するため、本Taskでは承認済みport契約と決定的test doubleで検証する。

## TASK-002-04 `GET /admin/generation-requests/{id}` をend-to-endで実装・検証する

**目的:** 管理者が記録済みの生成要求を安全に再読込みし、結果・失敗・中断時に残り得る`running`記録を画面で状態変更なしに確認できるようにする。

**関連要件:** REQ-011〜014、REQ-026、REQ-027、BR-016、NFR-001、NFR-003〜006、CON-003、CON-004

**作業内容:** `GET /admin/generation-requests/{id}` を、read guard、UUID検証、共通永続化基盤を用いた要求・attempt・candidate metadataのread-only取得、安全なresource/error変換、結果再表示画面まで一貫して実現する。GET固有のHTTP/画面/E2E検証を同じTaskに含める。

**受け入れ基準:**

- GETはread guardとUUID形式検証後に、request、attempt、candidate metadataだけを時系列順でread-only取得する。候補HTML本文、未検証出力、CSRF/session値、Codex内部ID・詳細errorは返さず、状態を変更しない。
- GETは200、UUID形式不正400、存在しない要求404、内部障害500を[HTTP契約](design/http-contract.md)どおりに返す。未認証401とsession store障害503はFEAT-001の共通error envelopeに従い、guard不通過・入力不正・不存在時に不要なDB読取りまたは外部生成を行わない。200はJSONと`Cache-Control: no-store`を返す。
- POSTのT2〜T4書込み失敗などで残り得る`running`を、状態変更や再開なしに安全に返す。completed_succeededではcandidate metadataを、completed_failedでは安全なfailureを、resource representationの全規定fieldとattempt順序で返す。
- 結果再表示画面は、POSTの201または422で得たrequest IDを結果画面へ渡し、初回表示・reloadのいずれでもGETで成功・失敗・`running`を表示する。要求不存在、無効ID、401/503、500を安全に案内し、成功・失敗・`running`から新規初回または修正文脈へ戻れる。候補HTML、candidate ID、source version ID、秘密値、Codex詳細を画面に露出しない。
- UI component、GET HTTP integration、E2Eで、成功・最終失敗・`running`の再読込み、400、404、500、未認証、`no-store`、read-only性、candidate本文非露出、結果画面のアクセシビリティを[テスト戦略](design/test-strategy.md)どおりに検証する。通常CI/E2Eは実Codex認証情報や外部ネットワークに依存しない。

**依存:** TASK-002-01（読取り対象となるmigration、request/attempt/candidate契約）、TASK-002-03（POST結果からの結果画面導線）、FEAT-001の管理read guard。

**対象外・注記:** GETは中断された生成の再開、取消、polling/通知を提供しない。候補または版の別origin隔離表示と共同E2EはFEAT-005が所有する。

## 実装へ委譲する事項

- 承認済み責務境界を満たす内部file path、package、component、class、function、query構造
- 承認済みcontractを変えない局所的なframework API、library API、test helper、dependency注入の構成

## 他Featureの責務

- FEAT-001: 管理session、CSRF、Origin、read/mutation guard、PostgreSQL migration runner。
- FEAT-003: 合格候補の採用、version・content・公開状態の所有、VersionSourceの実装。
- FEAT-005: 候補または版の別origin隔離表示、およびREQ-006の共同結合検証。
