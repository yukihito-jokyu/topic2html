# 作業セッション引継ぎ

このファイルはローカルセッション用であり、Gitへコミットしない。

## 担当作業

- 担当Issue: [#{{ISSUE_NUMBER}}]({{ISSUE_URL}})
- 作業ID: `{{TASK_ID}}`
- 作業名: {{ISSUE_TITLE}}
- 変更履歴の作業系列（Branch）: `{{BRANCH}}`
- 分離作業ディレクトリ（Worktree）: `{{WORKTREE_PATH}}`
- 作業開始の基準（Base ref）: `{{BASE_REF}}`
- 分離作業ディレクトリの起点Commit識別値（SHA）: `{{BASE_SHA}}`
- 次工程へ進む条件を満たしたCommit: `{{GATE_COMMIT}}`
- 計画固定版（Planning snapshot）: `{{PLANNING_SNAPSHOT}}`
- 書込み可能な場所と対象指定（Path／Glob）: {{OWNER_PATHS}}

## このファイルで使う用語

- GitHub Issue（課題記録）: 要求、担当範囲、完了条件、作業経過をGitHub上で管理する記録である。現在のTaskが何を満たすべきか確認するために必要である。
- Task Issue（実行対象の課題）: 独立した完了条件を持つ最小の作業単位。作業範囲を限定するために必要である。
- Branch（変更履歴の作業系列）とWorktree（分離作業ディレクトリ）: このTaskの変更と作業場所を他Taskから分けるために必要である。
- Base ref（作業開始の基準）とWorktree起点SHA（基準Commitの40文字識別値）: どの時点から作業を始めたか再現するために必要である。
- Gate通過Commit（次工程へ進む条件を満たした変更履歴）: 未確定の設計に基づいて後続作業を始めないために必要である。該当しないTaskでは「該当なし」とする。
- Planning snapshot（計画固定版）: Taskの分割、依存関係、担当範囲を特定のCommit時点で固定したもの。作業中の計画変更と開始時の条件を混同しないために必要である。
- Task Map（作業構成表）と接続台帳: Taskの担当・依存関係と、成果物を作るTaskから利用するTaskへの対応を記録した文書である。現在のTaskの位置と直接の後続Issueを確認するために必要である。
- 決定記録: 利用者が承認した変更内容、理由、影響範囲の履歴である。計画固定版より後の変更が正式承認済みか確認するために必要である。
- 書込み可能なPath／Glob（変更可能な場所と、その場所をまとめて表すパターン）: 担当外の成果物を変更せず、並行Taskとの競合を防ぐために必要である。
- Git状態と差分: 未保存・未Commitの変更、Commit候補への一時登録、現在のCommitと変更内容を示す情報である。未承認Commitや担当外変更がないか確認するために必要である。
- Commit（Gitへ変更履歴として保存する操作）、PR（変更の取込み依頼）、Merge（変更の取込み）: 変更の作成、確認依頼、正式な取込みという異なる段階を区別するために必要である。
- `human-progress`（人が更新する進捗欄）: 自動生成領域とは別に、実際のCommit、検証、PR、統合結果をIssueへ記録する場所である。実施前の予定を完了証拠として扱わないために必要である。

## 開始手順

1. このファイルを読む。
2. `$conduct-task-discussion` と `$explain-with-context` の `SKILL.md` を全文読む。
3. GitHub Issue本文、Issue #1、計画固定版、固定版と現在の接続台帳、Issueが指定する決定記録、直接の後続Issueを読む。
4. 現在のBranchとその最新Commit、依存Taskの完了証拠、Gateの通過証拠を確認する。
5. 書込み可能なPath／Glob以外を変更しない。
6. Task固有の完了条件と検証方法を確認してから作業を始める。

## Codex開始プロンプト（必須3スキル・版1）

```text
$manage-task-worktrees、$conduct-task-discussion、$explain-with-context を使って
.codex/task-session.local.md とTask Issueを読み、記載されたleaf Taskだけを実行してください。
ユーザーへの返答と成果物では前提知識、用語の意味、各要素が必要な理由を省略しないでください。
Commit前に別サブエージェントへ、Issue #1、直接の後続Issue、計画固定版と現在のTask Map・接続台帳、
Issueが指定する決定記録、作業セッション、Git状態と差分、変更成果物を読取り専用で照合させ、
矛盾・欠落・説明不足の修正指摘がなくなるまで対応してください。
依存関係、Gate、成果物Ownerに差異があれば作業を止めて報告してください。
ユーザーが明示的に承認するまでCommitしないでください。
```

## Commit承認前の完了確認

- Task固有の検証または見直しを実行する。
- 別サブエージェントの独立レビューで、Commit前に必要な修正指摘が残っていないことを確認する。
- 変更した場所が成果物の担当境界内であることを確認する。
- 利用者の明示承認前は、Commitも、次のCommit候補への一時登録も行わない。

## Commit承認後

- 利用者が明示的に承認した変更だけを、成果物の担当境界ごとに分けてCommitする。
- Commit後のpush（共有リポジトリへの送信）、PR作成、Issue更新は、利用者の依頼範囲と外部操作の承認を確認してから行う。
- PR作成やレビュー完了をMerge承認と解釈しない。利用者がMergeを明示的に承認するまでMergeしない。

## Merge後

- 実際のCommit、検証結果、PR、統合CommitをIssueの`human-progress`領域へ記録する。
- Merge前にWorktreeを削除しない。削除は統合済みか確認し、利用者の明示承認を得てから行う。
