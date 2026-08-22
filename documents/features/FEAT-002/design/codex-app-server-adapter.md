# Codex app-server adapter 契約

## 所有範囲

`CandidateGeneration`のCodex adapterは、Go Server内ではexecution brokerのclientとして動作する。brokerはGo Serverと異なるOS service accountで一つのgeneration attemptごとに一つのapp-server子processを所有し、stdio接続・thread・turnをattempt間で再利用しない。同時のHTTP要求はそれぞれ独立した実行を持ち、brokerは他要求のstdioへmessageを多重化しない。

brokerだけが、設定済み絶対実行可能ファイルを固定argv `app-server --stdio`で起動する。利用者入力、topic、instructions、version本文、環境由来の値をargv、cwd、環境変数へ追加しない。Go Serverのadapterはapp-serverの認証値を読み取らない。Go ServerのGoogle OAuth secret、PostgreSQL接続、CSRF保護鍵、管理session情報をbrokerまたはapp-server実行環境へ継承する起動は許可しない。

Go Serverとbrokerは、外部networkへ公開しないprivate local IPCだけで接続する。brokerはOS所有者・modeでGo Serverだけをclientに許可し、Go Serverのadapterには完成済みpromptを一attempt分だけ渡す操作と、HTML本文または`generation_unavailable`だけを返す結果を提供する。brokerは任意command・argv・cwd・環境・認証path・JSON-RPC ID・外部詳細errorをclientに受け取らせず返さない。promptはCodex認証情報、Google OAuth secret、PostgreSQL接続、CSRF保護鍵、管理session情報を含めない。

## 固定wireと実行制約

broker内のapp-server adapterは標準入出力上のJSON-RPC 2.0 v2だけを用いる。開始後、次の順で一回ずつ送る。request IDはbroker内だけで使い、永続化・HTTP・ログに出さない。

1. `initialize`のparamsを`{"clientInfo": {"name": "topic2html-server", "version": "<server build version>"}}`として送る。responseは同じrequest IDでなければならない。必須の`codexHome`、platform、user agentを検査に使わず直ちに破棄し、続けて`initialized` notificationを送る。
2. `thread/start`のparamsを`{"cwd": "<configured workdir>", "approvalPolicy": "never", "sandbox": "read-only", "ephemeral": true}`として送る。成功responseの`thread.id`をmemory上だけで保持する。
3. `turn/start`のparamsを`{"threadId": "<thread.id>", "input": [{"type": "text", "text": "<server-composed prompt>"}]}`として送る。成功responseの`turn.id`をmemory上だけで保持する。
4. 当該thread/turnの`item/completed`と`turn/completed`を読み、下記の選択規則を満たすまで待つ。

promptはServer側で作り、完全なHTML文書だけを回答し、tool・file操作・追加質問を行わないよう指示する。`thread/start`/`turn/start`にmodel、config、developer instructions、tool設定、network設定を追加しない。`read-only` sandboxと空の専用workdirにより、生成に不要な書込みとアプリ資産への作業directory経由の到達を許可しない。broker内adapterはapprovalを許可しない。app-serverからclientへのJSON-RPC request、approval request、command execution、file change、MCP/dynamic tool callを受信したら、そのattemptを`generation_unavailable`にする。

### 通知状態機械

`initialize` response前にはそのresponseだけを受け付ける。`initialized`送信後の`thread/start` response待ちと、`turn/start` response待ちでは、responseより先に届いた下記のscoped notificationを一時保留する。responseでthread IDまたはturn IDを得た時点で保留分を同じ状態機械に再投入し、IDが一致しないもの、許容されない段階のものは`generation_unavailable`にする。responseが届かなければ保留分を採用せず失敗にする。これにより、responseとnotificationの到着順を候補選択に依存させない。

`thread/start` response待ちで保留できるのは`thread/started`、`thread/status/changed`、`thread/settings/updated`であり、確定したthread IDと一致した後にだけ処理する。`turn/start` response待ちでは、これらに加え、下記のturn/item notification、`thread/tokenUsage/updated`、`error`を保留できる。`TurnStartResponse`はthread IDを返さないため、broker内adapterは送信済み`turn/start.params.threadId`を当該turnのthread IDとして保持し、responseではrequest IDと`response.turn.id`だけを照合する。以後のnotificationでその保持thread IDとturn IDの両方を照合する。

`turn/start` response後は、次の非終端通知を受け付ける。いずれもthread IDとturn IDが保持値に一致しなければ`generation_unavailable`にする。

- `turn/started`。
- `item/started`。item typeは`userMessage`、`reasoning`、`agentMessage`だけを許容し、item IDとtypeをmemory上で記録する。
- `item/agentMessage/delta`。先行する`agentMessage` item IDと一致するdeltaだけを許容する。delta本文は候補に使わない。
- `item/reasoning/*`のdelta。先行する`reasoning` item IDと一致するものだけを許容する。
- `item/completed`。先行する同じitem ID・typeの終端通知だけを許容する。`agentMessage`以外のcompleted itemは候補に使わない。
- `turn/completed`。agentMessageの`item/completed`後に一度だけ許容する。

`thread/status/changed`は同じthread IDだけを受け付け、`systemError`なら失敗、`notLoaded`/`idle`/`active`は採用せず破棄する。`thread/settings/updated`と`thread/tokenUsage/updated`はIDを照合して破棄する。`error`は同じthread/turn IDだけを受け付け、`willRetry = true`ならerror本文を破棄してterminal通知を待ち、`willRetry = false`なら直ちに`generation_unavailable`にする。これらのtelemetry/errorはHTML候補、DB、HTTP応答、安全ログに保存しない。

`item/started`でcommand execution、file change、MCP/dynamic tool call等の非許可itemを受けた時点、またはその関連deltaを受けた時点でそのattemptを失敗にする。thread/turnを持たない既知のwarning・config更新等のglobal notificationは無視する。上記以外のscoped notification、未知のmessage、JSON-RPC不正、response ID不正、順序不正は現在のattemptを`generation_unavailable`にする。thread/turn IDと外部エラー本文は直ちに破棄し、安全ログにも出さない。

## HTML出力の選択規則

一attemptからHTML本文を採用できるのは、以下をすべて満たす場合だけである。

- `turn/completed`が、保持中のthread IDとturn IDに一致し、`turn.status`が`completed`である。
- 同じthread IDとturn IDの`item/completed`に、`item.type = "agentMessage"`であるitemがちょうど一件ある。
- そのitemの単一`text`値が空でない。
- 上記以外のagentMessage、tool実行、client request、通信異常、途中終了がない。

採用本文はその一件の`item.text`を連結・整形・code fence除去せず、そのままHTML validatorへ渡す。agentMessageが0件または複数、空・非text出力、turnの`failed`/`interrupted`、完了通知前のstdio終了は`generation_unavailable`である。本文が存在してもHTML validatorで不合格なら`invalid_html`である。tool結果、reasoning、turn response内のitem配列はHTML候補に採用しない。

## 待機・切断・終了

NFR-006に従い、brokerが起動した正常な子processには完了時間のdeadlineを設けない。brokerはHTTP clientの接続状態を受け取らず、admit後は子processの終端結果を一回だけ返し、registry cleanupまでを完了する。HTTP client切断後もServerがgenerationを継続し応答を書かない契約は、[POST operation](operations/create-generation-request.md)の責務である。

### shutdown admissionの論理契約

brokerはattempt admission gate、稼働中attempt registry、app-server process groupを所有する。gateは`open`、`closing`、`closed`の一方向状態であり、brokerの一つの直列化点だけが`admit next attempt`と`begin shutdown`を処理する。Go Serverはgate・registry・processの状態を更新しない。Serverのshutdown flagは新規IPC要求を抑止する補助であり、競合の正しさはbrokerの直列化点だけで保証する。Server側のretry、永続化、HTTP終端は[生成operation](operations/create-generation-request.md)を正本とする。

| 直列化順 | brokerの規則 |
| --- | --- |
| `begin shutdown`が先 | gateを`closing`へ遷移し、以後のadmissionを`shutdown_rejected`として返す。processを起動せず、registryにも追加しない。 |
| `admit next attempt`が先 | gateが`open`の間にattemptをregistryへ登録してからprocess groupを開始する。後続のshutdownはそのregistry entryをcleanup対象とする。 |
| `closing`中 | 新規admissionをすべて拒否し、registryが空になるまでcleanupを継続する。空になった後に`closed`へ遷移する。 |

各admit済みattemptはregistryに一つだけ存在し、normal resultまたはshutdown中断結果のいずれか一つだけを返してから削除する。`shutdown_rejected`は外部interactionを開始していないため、brokerはprocessもregistry entryも作らない。

broker/adapterはadmission、process、結果分類だけを所有し、request・attempt・candidateのDB書込み、T4の選択、retry判断、HTTP応答を実行しない。`shutdown_rejected`は「admitされなかった」安全な制御結果であり、brokerがfailed attemptを合成してはならない。

shutdown開始時にgateを`open`から`closing`へ原子的に遷移させる。正常完了、抽出失敗、stdio異常、またはServer shutdownでbrokerがattemptを終了するときは、次の順序で子process groupを必ずreapする。

1. stdinをcloseし、新たなJSON-RPC requestを送らない。
2. 子processの終了を最大5秒waitする。
3. 未終了なら子process groupへterminateを送り、最大5秒waitする。
4. なお未終了ならkillを送り、終了するまでwaitしてreapする。

shutdownで中断した稼働attemptは一回だけ`generation_unavailable`としてGo Serverへ返す。brokerは結果を返す前に子processをreapし、結果受領後のretry、永続化、HTTP終端を待機・観測しない。この5秒はcleanupの猶予だけであり、正常なgenerationの完了上限ではない。

## adapter contract test

決定的test doubleは、private IPCの許可client・入力/出力制限、起動単位、`initialize`→`initialized`→`thread/start`→`turn/start`の送信順、固定params、request ID・thread response ID・turn response ID・item IDの一致、`turn/start` responseからthread IDを読まないこと、開始response前に届く保留notificationの再照合、candidate選択、admission gate、close/wait/terminate/kill/reap順を観測できなければならない。少なくとも、正常な単一agentMessage（started/delta/completedを含む）、複数agentMessage、空agentMessage、tool item、初期化response不正、thread/turn response前後のnotification、telemetry、分類可能/分類不能error、turn失敗、notification IDまたは順序不一致、stdio途中終了、shutdownとadmissionの競合でclose先行ならprocess/registry entryなしで`shutdown_rejected`、admit先行なら一回だけ中断結果とgroup cleanupを再現する。DB行、T4、retry、HTTP client切断後のServer継続、HTTP status、DB書込み失敗は観測・検証対象に含めない。
