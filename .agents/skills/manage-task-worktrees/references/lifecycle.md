# Worktreeライフサイクル

## 前提

- Primary checkoutのリポジトリ直下から実行する。
- `git`、`gh`、`code`を利用でき、`gh auth status`が成功する。
- `/.worktrees/`と`/.codex/task-session.local.md`が`.gitignore`に登録されている。
- 実行対象は`Lx-My-Sz`形式のleaf Task Issueである。
- 固定Planning snapshotのTask Mapと接続台帳を依存DAG・Gate・Ownerの正本とし、GitHub Issue本文を実行用の同期コピーとして照合する。差異があれば開始しない。
- GitHub Issue本文のTask ID Marker（そのIssueが実行対象Taskであることを示す識別欄）を確認する。Task ID Markerは、進捗管理だけを行う親Issueを誤って実行しないために必要である。
- 基準refと作成・再開するworktreeの双方に、`manage-task-worktrees`、`conduct-task-discussion`、`explain-with-context`の3つのSkillが存在する。開始プロンプトだけが存在し、実際の手順書を読めない状態を防ぐために必要である。

## 状態遷移

```text
未作成
  └─ plan合格
       └─ start
            └─ 作業中
                 ├─ open/status
                 └─ finish確認
                      └─ Commit・push・PR・Merge
                           └─ remove
                                └─ worktree削除済み
```

`plan`と`status`は読取り専用である。`start`はbranch（変更履歴の作業系列）、worktree（分離作業ディレクトリ）、ローカル引継ぎファイルを作成し、既定ではVS Codeを開く。既存worktreeを再開する場合は、質問・回答・判断履歴を上書きせず、不足する必須Skillの開始プロンプトだけを追記する。`finish`は完了判定材料を表示するだけでCommitやpushを行わない。`remove`はMerge済みで変更が残っていないworktreeだけを削除する。

## 基準ref

通常は`origin/main`を使用する。次の場合は`--base`を必須とする。

- Design Freeze Gate通過Commitから複数Taskを並行開始する。
- 依存Taskを統合した専用integration refから開始する。
- `origin/main`へ未統合だが、承認済みの共通起点Commitがある。

依存Issueはclosedだけでは不十分である。Issueのhuman-progress領域に統合Commitを記録し、そのCommitが基準refに含まれることを確認する。

Gate依存があるTaskは`--gate-commit <sha>`を指定する。スクリプトはGate Commitが基準refに含まれることを確認する。Gateの意味的な合格判定は行わない。

## 開始

leaf Taskを直接指定された場合は、最初にdry-run（worktreeを作らず開始条件だけを確認する試行）を行う。実際の作成前に、誤ったIssue、基準ref、依存関係を発見するために必要である。

```shell
rtk bash .agents/skills/manage-task-worktrees/scripts/manage_worktree.sh plan 28
```

表示内容を確認後、外部アクセスとGUI起動の承認を得て開始する。

```shell
rtk bash .agents/skills/manage-task-worktrees/scripts/manage_worktree.sh start 28
```

VS Codeを後から開く場合:

```shell
rtk bash .agents/skills/manage-task-worktrees/scripts/manage_worktree.sh start 28 --no-open
rtk bash .agents/skills/manage-task-worktrees/scripts/manage_worktree.sh open 28
```

## 初回起動準備

leaf Issueの機械識別欄、基準refの必須Skill、またはignore規則がない場合は、Taskの依存不成立ではなく、Issue作成工程とworktree管理導入工程の接続漏れとして扱う。利用者が内部形式を知っている前提で停止理由だけを返さず、次の順に復旧する。

1. Planning snapshotと実行baseを区別する。Planning snapshotはTask構成、依存、Gate、Ownerを固定したCommitであり、実行baseはTask worktreeを開始するCommitである。実行baseはPlanning snapshotを履歴に含み、さらにworktree管理用Skillとignore規則を含んでよい。
2. 現在のTask Map、議論記録、接続台帳を含む既存Commitを確認する。現在の計画資料に、そのCommit以後の未Commit差分がなく、Issue本文とも一致する場合は、新しい計画Commitを作らず既存CommitをPlanning snapshotとして再利用する。単に最新Commitという理由だけで選ばない。
3. GitHub Issue更新が利用者の依頼範囲に含まれる場合は、Task ID Marker、Planning snapshot SHA、固定リンクをIssueへ同期する。含まれない場合は、必要な外部変更と影響を説明して承認を得る。
4. 実行baseに必須Skill、`/.worktrees/`、`/.codex/task-session.local.md`のignore規則がなければ、管理機能導入用の分離branchで補う。このbootstrap branchはTask成果物のOwner branchではない。Task用`start`がまだ利用できない初回に限り、利用者の承認を得てPrimary checkoutから通常の`git worktree add`で隔離worktreeを作成できる。
5. 既存Planning snapshotを再利用できない場合だけ、同じ管理用branchでsnapshot候補も作る。Commit、統合、push、Issue更新はそれぞれの承認境界を守る。
6. 管理機能導入Commitを含むrefを実行baseとし、Planning snapshotがその履歴に含まれることを確認する。
7. Primary checkoutから`plan`を再実行し、通常の依存・Gate・Owner監査へ戻る。Task用worktreeを手動作成して検査を迂回しない。

「Task成果物を利用者の承認前にCommitしない」という規則と、開始前のPlanning snapshotは区別する。ただし、Planning snapshotであっても未承認のCommitを自動作成してよいという意味ではない。既存Commitの再利用を先に検討し、不要なCommitを要求しない。

同じTaskのbranchとworktreeが正しく存在する場合、`start`は破壊せず再開として扱う。別Pathで同じbranchがcheckoutされている場合は停止する。

## 親Issueからleafへの展開

tracking Issueを指定された場合はworktreeを作成せず、次の順で実行候補を解決する。

1. 指定IssueのTask ID、Planning snapshot、直下の子Issueを読む。
2. `Lx`または`Lx-My`の子を同じ方法で再帰的に辿り、`Lx-My-Sz`の子孫leafを列挙する。子IssueはIssue本文のgenerated領域にある「直下の子Issue」だけから辿る。
3. 各子のTask ID、親ID、Planning snapshotが親および固定Task Mapと一致することを確認する。差異があれば開始しない。
4. 固定Planning snapshotのTask Mapと接続台帳を読み、着手依存、完了・Merge依存、Gate、Owner Path、Merge順を復元する。現在のIssue本文だけから新しい依存を推測しない。
5. 依存Issueの状態、human-progressの統合Commit、baseへの包含、Gate Commit、既存worktreeを確認し、現在開始できるleafをready frontierとする。
6. ready frontier以降は、先行Taskの統合により解放される条件付きの将来waveとして示す。先行統合前に将来waveのbase SHAを確定しない。
7. 現在waveの候補だけをユーザーへ提示し、選択後に各leafの`plan`を実行する。tracking Issueへ`plan`や`start`を実行しない。

提示表には次を含める。

| 項目 | 内容 |
| --- | --- |
| Task | Issue番号、Task ID、Task名 |
| Ready | 現在開始可能か、未充足条件 |
| Evidence | 依存Issue、統合Commit、Gate Commit |
| Base | base ref／SHA。将来waveは解放条件だけを示す |
| Owner | 書込みPath／Glob、共有資産Owner |
| Execution | 並行可能Task、直列化理由、Merge順、次wave解放条件 |

Issue #3のPlanning snapshot `9bf911b9122bf4ec51ec48312cc310de29bfcdff`では、次のwaveになる。これは動線例であり、実行時は必ずIssueと固定snapshotを再取得する。

```text
Wave 1: #28
Wave 2: #29 + #30
Wave 3: #31
Wave 4: #32
Wave 5: #33
Wave 6: #34 + #35
Wave 7: #36
Wave 8: #37
Wave 9: #38
```

Wave 2は#28の統合Commit、Wave 6は#33の統合Commitをそれぞれ共通起点にでき、各組のOwner Pathが分離されているため並行可能である。各組内のMerge順は任意だが、両Taskを統合したCommitを次waveのbaseにする。それ以外は直接依存DAGに従って直列化する。

## waveの承認と作成

現在waveの全leafで`plan`が合格したら、次をまとめて提示して承認を得る。

- branch、worktree、handoff
- base ref／SHA、依存統合Commit、Gate Commit
- Owner Path、共有資産、既存worktreeとの競合監査結果
- 並行可能な組、直列化理由、Merge順、次waveの解放条件
- GitHubアクセス、fetch、ローカル作成、VS Code起動の有無

複数leafをまとめて作る場合は、承認後にleafごとに`start --no-open`を実行し、必要なウィンドウだけ`open`してもよい。`start --no-open`もbranch、worktree、handoffを作成する変更操作である。後続waveは前提Merge後に再展開・再`plan`し、別途承認を得る。

再開時に`.codex/task-session.local.md`が既にある場合、質問・回答・判断履歴を上書きしない。3つの必須Skillを指定する開始プロンプトがない場合だけ末尾へ追記する。worktree自体に必須Skillがない場合は、基準refをそのTask branchへ安全に統合するまで再開しない。

## 新しいCodexセッション

VS Code内でCodexを手動開始し、次のように依頼する。

```text
$manage-task-worktrees、$conduct-task-discussion、$explain-with-context を使って
.codex/task-session.local.md とTask Issueを読み、記載されたleaf Taskだけを実行してください。
ユーザーへの返答と成果物では前提知識、用語の意味、各要素が必要な理由を省略しないでください。
Commit前に別サブエージェントへ、Issue #1、直接の後続Issue、固定Planning snapshot、
変更成果物の矛盾・欠落・説明不足をレビューさせ、修正指摘がなくなるまで対応してください。
依存関係、Gate、成果物Ownerに差異があれば作業を止めて報告してください。
ユーザーが明示的に承認するまでCommitしないでください。
```

新しいセッションは、着手前に次を確認する。

1. Task IDとIssue番号が一致する。
2. 現在branchと引継ぎbranchが一致する。
3. HEADと引継ぎのworktree起点SHAが整合する。
4. 書込みPathがIssueの単一Owner境界内である。
5. 依存・GateのEvidenceが引継ぎとIssueに存在する。
6. `$conduct-task-discussion` と `$explain-with-context` が開始プロンプトに明記されている。
7. 固定版と現在の`docs/task-connections.md`、Issueが指定する決定記録から直接の後続Issueを特定できる。

## 並行作業

同じGate Commitから別worktreeを作る場合も、次を満たす必要がある。

- branch名とworktree PathがTaskごとに一意である。
- 書込みPath／Globが重ならない。
- 共有Registry（複数Taskが利用する登録一覧）、Lockfile、生成物、共通Fixtureは承認済み単一Ownerだけが変更する。複数Taskによる同時変更と競合を防ぐためである。
- Merge順があるTaskは、worktree作成を並行化しても確定・Mergeを直列化する。
- 上流契約をconsumer（その契約を利用する側）のTaskから変更しない。契約を定義するTaskと利用するTaskの責任を混在させないためである。

並行可能と提案するのは、次をすべて満たすTaskだけとする。

- Task間に着手依存辺がない。
- 同じ依存統合CommitとGate Commitを共通起点にできる。
- Owner Path／Globが交差しない。
- Schema、Migration、共通Interface、Registry、DI、Lockfile、生成物、共有Fixtureが単一Ownerまたはread-onlyである。
- 各leafの`plan`が個別に合格する。
- Merge順と、全Taskを統合した次waveのbase作成条件が確定している。

スクリプトは正規化可能な完全一致、親子Path、単純な`/**` prefixを既存worktreeと比較する。複雑Glob、除外規則、共有資産の意味的な競合は完全判定できないため、Codexが候補wave内と既存worktreeを横断監査する。自動検査合格を「競合なし」の根拠にしない。

Path競合はスクリプトだけで完全判定しない。CodexがIssueの「成果物と所有」とTask Mapを読み、意味的に確認する。

## Commit承認前の完了準備

```shell
rtk bash .agents/skills/manage-task-worktrees/scripts/manage_worktree.sh finish 28
```

出力されたbranch、HEAD、変更Path、Owner PathをIssueと照合する。その後、Task固有テストと別サブエージェントによる独立レビューを実行する。利用者が明示的に承認する前は、Commit候補への一時登録もCommitも行わない。

## Commit承認後

利用者が明示的に承認した変更だけを、成果物Ownerごとに分けてCommitする。push、PR作成、Issue更新はCommit承認と同一の操作ではないため、利用者の依頼範囲と外部操作の承認を別に確認する。

## PR作成・統合後の記録

PR作成やレビュー完了はMergeの承認ではない。利用者がMergeを明示的に承認するまでMergeせず、承認後にだけ実行する。

Issueへ少なくとも次を記録する。

- worktree起点SHA
- 実装または設計Commit
- 実行した検証Commandと結果
- PR
- 統合Commit
- Gate／Releaseに渡すEvidence
- 未解決事項の有無

実施前の予定を完了証拠として記録しない。PRがMergeされた後に、実際の統合Commitを記録する。

## 削除

PRのMergeとIssue Evidence記録後に実行する。

```shell
rtk bash .agents/skills/manage-task-worktrees/scripts/manage_worktree.sh remove 28 \
  --merged-into origin/main \
  --confirm
```

ローカルbranchも削除する場合だけ`--delete-branch`を付ける。スクリプトはdirty状態、Merge未完了、Path不一致で停止する。`--force`は提供しない。

## 障害時

- Issue Marker（実行対象Taskを識別するIssue本文の印）がない: tracking Issue（複数Taskの進捗だけをまとめる管理用Issue）の可能性があるため開始しない。実際の成果物を持たない親Issueで作業を始めないためである。
- leaf TaskであることをIssue本文とTask Mapから確認できるのにIssue Markerがない: [初回起動準備](#初回起動準備)へ戻り、Issue作成工程との接続を修復する。
- 基準refに必須Skillまたはignore規則がない: [初回起動準備](#初回起動準備)へ戻り、管理機能導入用branchを実行baseへ統合する。Task worktreeの手動作成では迂回しない。
- Task IDがleafでない: 親Taskは進捗管理用なので開始せず、「親Issueからleafへの展開」を実行する。
- 依存Commitがない: 依存IssueへEvidenceを記録する。
- Gate Commitがない: Gate確認セッションへ戻る。
- 基準refにCommitがない: integration順を修正し、refを更新する。
- Pathが既に存在する: 登録済みworktreeとbranchを確認し、推測で削除しない。
- `code`がない: `start --no-open`を使用し、利用可能なVS Code起動方法をユーザーと決める。
