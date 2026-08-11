---
name: Leaf task
about: 承認済みTask Mapのleafを実行する
title: ''
labels: ''
assignees: ''
---

<!--
タイトル形式: [Lx-My-Sz] タスク名

このIssueは、承認済みTask Mapに存在するleafを実行するためのものです。
Issue本文から新しい依存、成果物、Path、設計判断を追加しないでください。
Task Map、決定記録、Gate記録との不一致を発見した場合は作業を止め、
正本を再承認・更新してからこのIssueを同期してください。

自動作成時は仕様部分をgenerated-content Markerで管理し、実施チェック、
着手・検証Evidence、PRはhuman-progress Marker内へ分離します。
再同期はhuman-progress内を変更しません。
-->

## タスク情報

- Task ID:
- 親Task ID:
- タスク名:
- タスク種別: <!-- design / implementation / evaluation -->
- Planning snapshot commit SHA:
- Task Map固定リンク: <!-- https://github.com/yukihito-jokyu/topic2html/blob/<SHA>/docs/task-map.md#... -->
- 原典Issue: https://github.com/yukihito-jokyu/topic2html/issues/1
- 関連する原典章:
- 関連する決定ID:

## 目的

<!-- このleafが完成させる単一の到達状態 -->

## 原典との差分

- 固定入力:
- 原典で決定済みだが未実施の成果物:
- このleafで決める未確定事項:
- 選び直さない事項:

## 実施内容

- [ ]
- [ ]

## 成果物と所有

| 項目 | 内容 |
| --- | --- |
| 主成果物 | |
| 書込み可能なPath／Glob | |
| 単一Owner | <!-- 原則としてこのTask ID --> |
| read-only入力 | |
| 共有資産と単一Owner | |
| Gate通過記録 | |

書込み可能なのは、上表の所有Pathと承認済みの明示例外だけです。それ以外のPathは変更しません。

## 完了条件

- [ ]
- [ ]
- [ ] 主成果物が指定Pathに存在する
- [ ] 単一Ownerと書込みPathの境界を守っている
- [ ] Task固有の検証結果またはReview evidenceをIssueへ記録した
- [ ] 未解決の契約差異とTBDがない

## 対象外

-
-

## 依存関係

| 種別 | Task／Milestone | 必要な成果物・状態 |
| --- | --- | --- |
| 着手依存 | | |
| 完了・Merge依存 | | |
| Gateへの入力 | | |
| Gate通過依存 | | |
| Release条件 | | |

- 後続接続の固定リンク: <!-- https://github.com/yukihito-jokyu/topic2html/blob/<SHA>/docs/task-map.md#... -->

<!--
依存関係はPlanning snapshotのTask MapとGate記録から転記します。
推移的依存、親roll-up、新しい依存判断を追加しません。
後続接続は派生文書への参照に留め、Issue側で編集しません。
-->

## 着手判定

| 確認項目 | 結果・Evidence |
| --- | --- |
| Planning snapshot SHAとTask Map固定リンクが一致する | |
| 着手依存TaskのMerge commit | |
| 必要なGateの通過記録、またはGate前Taskであること | |
| worktree起点SHA | |
| 必須値に未解決TBDがない | |
| 並行Taskと書込みPathが競合しない | |

- [ ] 上表を確認し、このTaskは着手可能である

## worktree・Merge

| 項目 | 内容 |
| --- | --- |
| planning baseline SHA | |
| worktree起点SHA | |
| Branch | |
| 所有Path／Glob | |
| 共有物と単一Owner | |
| 並行可能Task | |
| 直列化するTaskと理由 | |
| Merge前提 | |
| Merge順 | |
| 統合先 | |

## 検証

| 種別 | 方法・Command | 合格条件 | Evidence |
| --- | --- | --- | --- |
| 静的確認 | | | |
| Task固有テスト／Review | | | |
| 契約・統合確認 | | | |
| 後続評価 | <!-- L6等への参照。ここでは重複実施しない --> | | |

<!--
該当しない項目は削除せず、次の形式で理由を記載します。
該当なし（理由: ...）

Task ID、目的、成果物、OwnerはIssue作成時から必須です。
Gate待ちの本番Path、起点SHA、PR等は作成時のみTBDを許容しますが、
Ready判定時にはTBDを残しません。
-->

## 差異を発見した場合

- [ ] 作業を停止する
- [ ] Issue内で新しい依存、Path、設計判断を決めない
- [ ] Task MapまたはGate記録の修正案を議論記録へ残す
- [ ] 必要な再承認後、Planning snapshotとIssueを同期する

## 関連Issue・PR

- 親Issue:
- Blocked by:
- 関連PR:
