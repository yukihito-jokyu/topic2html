# DEC-FEAT-003 — Codex実行資格情報をGo Serverから分離する運用境界

状態: **承認済み（L3、2026-08-16）**

## 課題

現在のGo Serverは、PostgreSQL接続、Google OAuth client secret、CSRF保護鍵を同じ実行環境から読み取る。OSの通常の子process起動では、この環境と実効UIDが継承されるため、既存の「専用service accountのapp-server実行環境だけがCodex認証を扱う」という記述だけでは、子processにアプリ秘密を渡さないことも、Go ServerにCodex認証を読ませないことも保証できない。

このDecisionはCodexアカウントの発行・権限・課金主体を決めない。ServerとCodex実行環境をどのOS/プロセス境界で分離するかだけを決める。

## 選択肢

### A. 専用service accountのローカル実行brokerを設ける

- Go Serverとは異なるOS service accountで、Codex認証情報、Codex用home、空workdir、app-server子processを所有する実行brokerを運用する。
- Go Serverは、認可済みの生成入力だけをprivate local IPCでbrokerへ渡す。brokerは固定実行可能ファイル・固定argv・allowlistした最小環境だけでapp-serverを起動し、HTML本文または安全な失敗分類だけを返す。
- Google OAuth client secret、PostgreSQL接続文字列、CSRF保護鍵、管理session情報はbrokerの環境、workdir、IPC payloadへ渡さない。Codex認証情報はGo Serverの環境、設定、ログ、DB、Browserへ渡さない。
- IPC endpointは同一hostのprivate endpointとし、OS所有者・modeでGo Serverだけをclientとして許可する。brokerは外部networkからlistenせず、任意command・argv・cwd・環境・認証pathをclientに指定させない。
- brokerが実行可能ファイルとworkdirを起動時に検証し、Go Serverはprivate IPC endpointだけを起動時に検証する。これはDEC-FEAT-002の「Go Serverが実行可能ファイルとworkdirを直接検証する」配置を置き換えるが、固定argvとfail-fastの不変条件は維持する。

### B. Go Serverが固定の権限昇格wrapperを直接起動する

- Go Serverは固定pathのwrapperだけを起動し、wrapperが別service accountへ切り替えてCodex認証を持つ最小環境でapp-serverを開始する。
- wrapperの権限、入力検証、process group、ログ、更新手順を別途運用・監査する。

## 推奨と根拠

利用者は**Aを承認した。** 異なるOS identityとprivate IPCにより、Go Serverの秘密環境とCodex認証環境を双方向に分離できる。Bはprivilege-escalation binaryをGo Serverの攻撃面に置き、入力・更新・監査の失敗がservice account境界を破るため採用しない。

## 承認後の影響

- `CandidateGeneration`のCodex adapterはbroker clientとなり、brokerが子processとCodex認証を所有する。runtime設定、adapter契約、smoke test、運用配備にprivate IPCとOS identityの検証を追加する。
- brokerは実行可能ファイルとworkdirを起動時に検証し、Go Serverはprivate IPC endpointを起動時に検証する。通常CIは決定的test doubleを使い、実Codex認証情報を使わない。
