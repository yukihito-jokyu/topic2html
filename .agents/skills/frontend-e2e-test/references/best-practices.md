# Playwright確認事項

- role、label、text、test idの順にlocatorを選ぶ。
- auto-waitingとweb-first assertionを使い、固定sleepを使わない。
- テストごとに独立したBrowser contextと可変状態を使う。
- CIでは`forbidOnly`、最小browser、retry時のtrace、`failOnFlakyTests`を設定する。
- trace、video、report、test-resultsはGit管理しない。
