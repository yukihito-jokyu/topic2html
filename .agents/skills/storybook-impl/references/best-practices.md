# Storybookの確認事項

- 依存追加・更新前に公式互換要件を確認する。
- Storyを対象部品の近くに置き、Props、`args`、`fn()`で状態を再現する。
- a11y検査は原則errorにする。
- 実APIや外部visual regressionサービスへ無断送信しない。
- static buildとbrowser testを、通常のfrontend build・静的解析とともに実行する。
