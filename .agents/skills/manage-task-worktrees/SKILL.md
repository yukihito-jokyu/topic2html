---
name: manage-task-worktrees
description: GitHub tracking Issueから実行可能なleaf Taskを展開し、依存DAGと書込みPathから並行waveを提案したうえで、Taskごとに隔離したGit worktree、ブランチ、VS Code、Codexセッションを作成・再開・完了・削除まで管理する。親Issueを起点に作業を始める場合、依存Task・基準Commit・Path競合を確認して並行作業を計画する場合に使用する。
---

# Task Worktree管理

Task Issueを1つのブランチ、worktree、VS Codeウィンドウ、Codexセッションへ対応付ける。

## 用語と前提知識

- Task Issue（実行対象の課題）とleaf Task（これ以上分割しない最小Task）: 独立した完了条件を持つ一つの作業単位である。親Taskを誤って実行対象にせず、変更範囲を限定するために必要である。
- Primary checkout（基準となる作業ディレクトリ）: `.git`の共通管理情報を持ち、他のworktreeを作成・削除する元の作業場所である。worktree管理を一か所へ集約するために必要である。
- branch（変更履歴の作業系列）とworktree（分離作業ディレクトリ）: 一つのTaskの変更履歴と実ファイルを他Taskから分ける仕組みである。未完了変更の混在を防ぐために必要である。
- ref（Commitを指す名前）、Base ref（作業開始の基準）、integration ref（複数Taskを統合した基準）: どのCommitを作業の起点や統合結果として使うか表す名前である。依存Taskを含む正しい時点から開始するために必要である。
- Commit（Gitへ保存した変更履歴）、SHA（Commitの識別値）、HEAD（現在のbranchが指す最新Commit）: 作業の開始・現在・統合済みの各時点を一意に照合するために必要である。
- Gate（設計上の確認記録）: 設計上の判断や確認事項を記録する項目である。worktreeの開始可否はGateではなく、明示された依存Task・基準ref・Owner Path・Path競合で判断する。
- 成果物Owner（変更責任の所在）とPath／Glob（変更場所と複数場所を表す対象指定）: Taskが変更してよいファイルや設計領域を示す。並行Task間の競合と責任混在を防ぐために必要である。
- Evidence（実施を裏付ける証拠）と`human-progress`（人が更新するIssue進捗欄）: 実際のCommit、検証、PR、統合結果を確認・記録するために必要である。予定を完了実績として扱わない。
- dirty／clean（未Commit変更がある／ない状態）: worktreeを安全に削除できるか区別する状態である。未保存の成果物を失わないために必要である。
- push（共有リポジトリへの送信）、PR（変更の取込み依頼）、Merge（変更の取込み）: ローカル履歴の作成、共有、確認依頼、正式取込みを分けて扱うために必要である。
- GUI（画面操作）、VS Code（作業画面を開く編集環境）、Codexセッション（Codexとの一つの作業会話）: Task専用の作業場所と会話を利用者が開始・再開するために区別する。
- Schema（データ構造）、Migration（保存データ構造の移行処理）、Interface（プログラム間の操作境界）、Registry（共有する登録一覧）、DI（必要な処理を外部から渡す依存性注入）、Lockfile（依存部品の版固定ファイル）、Fixture（検証用の固定データ）: 複数Taskが同時変更すると競合しやすい共有成果物である。単一Ownerまたは直列Mergeが必要か判断するために区別する。

## 必須手順

1. 変更操作の前に[ライフサイクル](references/lifecycle.md)を全文読む。
2. 指定Issueがtracking Taskなら、`plan`や`start`を実行せず、Issue本文と現在のTask Mapから子孫leafを再帰的に展開する。
3. leafの依存DAG、基準ref、Owner Path、既存worktreeを照合し、Gateは参考情報として併記して、現在のready frontierと将来waveをユーザーへ提案する。
4. 基準refの必須Skillまたはignore規則が欠けている場合は、初回起動準備の未完了として扱う。[ライフサイクルの初回起動準備](references/lifecycle.md#初回起動準備)に従い、管理機能を導入するbase Commitの要否を判断する。利用者に未説明の内部前提を補わせる質問だけを返して作業を止めない。
5. ユーザーが選んだ現在waveの各leafについてPrimary checkoutから `plan` を実行する。機械判定に加え、複雑Globと共有資産のPath競合を意味的に監査する。
6. wave全体の作成内容、並行・直列理由、Merge順、外部操作を示し、`plan`が合格した各leafの `start` を続けて実行する。worktree作成のための利用者承認待ちは挟まない。
7. VS Codeが開いたら、ユーザーへ各ウィンドウでCodexの開始と引継ぎファイルの読込みを依頼する。生成された開始プロンプトでは `$conduct-task-discussion` と `$explain-with-context` も必ず指定する。
8. 作業完了時は `finish`、Issue固有の検証、別サブエージェントの独立レビューを実行する。利用者の明示承認前はCommit候補への一時登録もCommitも行わない。
9. Commit承認後は成果物OwnerごとにCommitを分ける。push、PR作成、Issue更新は、利用者の依頼範囲と外部操作の承認を確認してから行う。CommitをPrimary checkoutのlocal `main`へ直接Mergeしない。
10. Mergeは、PR作成とは別の外部変更である。利用者がMergeを明示的に承認するまで実行せず、PR作成やレビュー完了をMerge承認と解釈しない。承認後はリモートの`main`へPRをMergeし、Primary checkoutでは `git pull origin main` で統合結果を取得する。
11. Merge後に限り、明示承認を得て `remove` を実行する。

公開GitHub Issueの読取りは認証なしのREST APIで行う。private Issueの読取り、push、PR、削除など認証または外部変更を伴う操作では、それぞれ既存の承認規則に従う。

## 親Issueからの開始

`Lx`と`Lx-My`はtracking Task、`Lx-My-Sz`はworktreeを持つleaf Taskとして扱う。tracking Issueを指定された場合は[ライフサイクルの親Issue展開](references/lifecycle.md#親issueからleafへの展開)を実行し、次を表で提示する。

- 現在開始できるready frontierと、依存統合後に解放される将来wave
- 各leafのIssue、Task ID、依存Evidence、Gateの参考情報、Owner Path、base ref／SHA
- 並行可能な組、直列化する理由、Merge順、次waveの解放条件

将来waveのbase SHAは先行Taskの統合前に確定しない。tracking Issue自身へ `start` を実行せず、ユーザーが承認した現在waveのleafだけを `plan`、`start` する。

## コマンド

Skillディレクトリを `<skill>` として、すべてPrimary checkoutから実行する。

```shell
rtk bash <skill>/scripts/manage_worktree.sh plan <issue-number>
rtk bash <skill>/scripts/manage_worktree.sh start <issue-number>
rtk bash <skill>/scripts/manage_worktree.sh status [issue-number]
rtk bash <skill>/scripts/manage_worktree.sh open <issue-number>
rtk bash <skill>/scripts/manage_worktree.sh finish <issue-number>
rtk bash <skill>/scripts/manage_worktree.sh remove <issue-number> --merged-into <ref> --confirm
```

通常の基準refは `origin/main` とする。Gateは設計上の参考情報としてIssueとTask Mapで扱う。Gateの未通過、記録不足、または記載差異を理由に`plan`、`start`、ready frontierを停止しない。

`plan`はworktreeを作成しない。単純なPath重複は検査するが、複雑Glob、除外規則、Schema、Migration、Interface、Registry、DI、Lockfile、生成物、共有FixtureはCodexがTask MapとIssueを意味的に監査する。`start --no-open`はVS Codeを開かず、worktreeと引継ぎファイルだけを作成する。

## 配置規則

Issue #28、Task ID `L1-M1-S1`の例:

```text
worktree: <repo>/.worktrees/issue-28-l1-m1-s1
branch:   task/issue-28-l1-m1-s1
handoff: <worktree>/.codex/task-session.local.md
```

`/.worktrees/`と`/.codex/task-session.local.md`はGit管理対象外にする。worktree内から別worktreeを作成しない。

## Codexへの引継ぎ

`start`はVS Codeを新規ウィンドウで開き、ローカル引継ぎファイルを表示する。Codex UIの開始はユーザーが手動で行う。新しいCodexは次を守る。

- `plan`は基準ref、`start`は基準refに加えて既存branchと対象worktreeに、`manage-task-worktrees`、`conduct-task-discussion`、`explain-with-context`が存在することを確認する。開始プロンプトが、存在しないSkillを指定する状態を許可しない。
- 既存worktreeの再開では`.codex/task-session.local.md`を上書きしない。質問・回答・判断履歴を保持し、必須Skillの開始プロンプトがない場合だけ追記する。
- `start`だけでなく`open`でも、worktree内の必須Skillと版付き開始プロンプトを検査する。`open`を使って検査を迂回させない。
- 引継ぎファイル、Task Issue、現在のTask Mapを最初に読む。
- `$conduct-task-discussion` と `$explain-with-context` を読み、利用者への返答、成果物、質問、作業セッション記録へ適用する。
- Commit前に別サブエージェントへ、Issue #1、直接の後続Issue、現在のTask Map、変更成果物の整合性を独立レビューさせる。
- Issueの単一Owner Pathだけを変更する。
- 意味判断、依存関係、Gate条件を作業セッション内で追加しない。
- 契約差異や未解決TBDを見つけた場合は実装を止め、元セッションへ戻す。
- 親Taskではなくleaf Issueだけを実行単位にする。

## 安全境界

- Task用worktreeを手動作成しない。通常の`git worktree add`を使えるのは、[初回起動準備](references/lifecycle.md#初回起動準備)で管理機能導入用branchを作る場合だけとする。
- dirtyなworktreeを削除しない。
- `--merged-into`へHEADが含まれないworktreeを削除しない。
- `--force`、`git reset --hard`、未確認のbranch削除を使わない。
- leaf依存Issueが未完了、統合Commitが未記録、または基準refに含まれない場合は開始しない。親依存Issueは、すべての子孫leafが完了Evidenceを持ち基準refに含まれる場合、親Issueのcloseや親専用の統合Commitを待たない。
- 並行Taskの書込みPathが重なる場合は開始せず、単一Ownerへ直列化する。
- 自動Path検査の合格だけで並行可能と断定しない。
- 未完了の将来wave用worktreeを先に作成して依存確認を迂回しない。
- 自動Merge、自動push、自動Issue closeは行わない。
