# DEC-FEAT-004 — shutdown時の生成停止と子process groupの責務

状態: **承認済み（L3、2026-08-16）**

## 課題

shutdown開始後に新規attemptを止めるには、request受理だけでなく、再試行直前のprocess起動を原子的に拒否するadmission gateが必要である。またapp-serverが孫processを作る可能性があるため、親PIDだけを終了しても子孫を残す可能性がある。既存設計は終了順序を示していたが、gateの所有者、in-flight requestの終端規則、process groupのOS契約が未確定だった。

## 選択肢

### A. shutdownでadmissionを閉じ、全実行groupを強制cleanupして要求を終端失敗にする

- 実行brokerがattempt admission gateと稼働中attempt registryを所有する。gateが`open`から`closing`へ遷移した後は、brokerもGo Serverも新しいattemptを起動しない。
- 各attemptは独立したOS process groupで開始し、brokerだけがgroup IDとstdinを保持する。cleanupはstdin close、新規JSON-RPC送信停止、5秒wait、groupへterminate、5秒wait、groupへkill、wait/reapの順とする。
- shutdownで中断した稼働attemptは一回だけ`generation_unavailable`として扱い、同一generation requestの後続retryは開始せず最終失敗へ収束する。Go ServerはHTTP受理停止後も、broker cleanupと終端記録を完了するまでDB等の依存を閉じない。
- DB終端記録が失敗した場合も子processはreap済みにし、既存のDB失敗規則どおり`running`記録を安全に残す。次回起動で再開はしない。

### B. 新規attemptだけ止め、開始済みattemptは通常完了までdrainする

- admission gateは閉じるが、実行中attemptはdeadlineなしで成功・通常retry・記録を続ける。
- Server停止は稼働attemptの終端まで待ち、process groupへのterminate/killは異常時だけに限る。

## 推奨と根拠

利用者は**Aを承認した。** shutdownが運用上いつ終わるかを管理でき、app-serverとその子孫を残さず、停止後のretry起動競合も防げる。BはNFR-006のdeadlineなしと組み合わさると、停止が無期限に延び、運用者が強制停止したときにcleanup・記録の一貫性を失うため採用しない。

## 承認後の影響

- broker/adapter contractにadmission gate、稼働attempt registry、OS process group、shutdownの終端規則を明記し、adapter contract testで起動競合、group signal、reap、安全な結果を検証する。終端記録はGenerationStore/POST operationの検証対象とする。

## L2設計補正 — shutdown先行時の記録

`attempt`はbrokerがadmitしregistryへ登録した外部interactionだけを指す。したがって`begin shutdown`がadmissionより先に直列化されて返る`shutdown_rejected`は、新しいprocessもattemptも作らない。このときGenerationStoreは、既存の`running` requestを安全な`generation_unavailable`で`completed_failed`へ更新するrequest-onlyのT4を提供する。admit済みattemptがshutdownで中断された場合だけ、failed attemptとrequestを同一transactionで終端化する。

これは、選択肢Aの「新規attemptを起動しない」と「中断した稼働attemptを一回だけ記録する」をそのまま適用するL2の論理補正である。HTTP応答、失敗分類、最大4回、または資格情報・process所有の承認済み契約を変更しないため、新たなL3/L4判断は不要である。
