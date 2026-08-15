---
name: skill-maintainer
description: Diagnose and repair problems in the planning orchestrator, planning skills, workflow policies/contracts, implementation-skill builder, or generated implementation skills. Use when a skill asks too many questions, produces poor artifacts, violates ownership, misses required analysis, chooses bad boundaries, triggers incorrectly, or repeatedly causes rework. Make minimal evidence-based changes and add regression scenarios. Do not hide product/design defects by changing skills.
---

# Skill Maintainer

## Goal

Skill運用中に見つかった不具合・摩擦・再現性の低さを、**症状ではなく原因に対して修正**する。

Skill自身をソフトウェアと同じように保守する。

## Trigger Examples

- 不要な質問が多くAIが頻繁に停止する
- Human Decision Point判定がおかしい
- 必要な設計分析を飛ばす
- 不要な設計図を常に作る
- Task粒度が毎回不適切
- SkillがOwnerではないArtifactを書き換える
- Orchestratorが誤ったSkillへルーティングする
- Implementation Skillが技術変更に追従しない
- Skill descriptionが広すぎ/狭すぎて誤発火する
- 同じ失敗が複数回発生する

## Inputs

可能な限り以下を集める。

- 問題の説明
- 問題が発生したSkill
- 実際の入力
- 実際の出力
- 本来期待した結果
- 関連Artifact
- `.ai/workflow/*.md|yaml`
- 既存 regression scenarios
- 会話ログや実行trace（利用可能な場合）

## Writable Scope

このSkillは修正目的で以下を編集可能。

- `.agents/skills/**/SKILL.md`
- Skill配下の `references/`, `scripts/`, `assets/`
- `.ai/workflow/decision-policy.md`
- `.ai/workflow/artifact-map.yaml`
- `.ai/workflow/implementation-handoff-schema.yaml`
- `.ai/skill-tests/**`

ただしProject Requirements / Design / Tasksを書き換えてSkill不具合を隠してはならない。

## Diagnostic Categories

問題をまず分類する。

### A. Activation / Description

- Skillが呼ばれない
- 関係ない場面で呼ばれる
- Skill同士のscopeが重複する

### B. Input Contract

- 必要情報を読んでいない
- Preconditionが過剰
- 正規ソースの優先順位が不明

### C. Procedure

- 手順が欠落
- 儀式的で冗長
- 自走できる処理で止まる

### D. Decision Policy

- L1/L2を人間へ聞く
- L3/L4を勝手に決める
- 影響範囲判定が曖昧

### E. Artifact Ownership / Handoff

- 別Skillの成果物を直接編集
- 入出力がつながらない
- 配置先が不統一

### F. Skill Boundary

- Skillが巨大すぎる
- Skillが細かすぎる
- 同じ責務が複数Skillへ重複

### G. Implementation Drift

- Tech Stack変更にGenerated Skillが追従していない
- 既存コードとSkill指示が矛盾

## Procedure

### 0. Structural Validation

リポジトリルートで、利用可能なら次を実行する。

```bash
python .agents/skills/skill-maintainer/scripts/validate_skillset.py
```

構文・Owner参照の問題を、モデル挙動の問題と混同しない。

### 1. Reproduce / Evidence

問題を再現できる入力条件を特定する。
再現できない場合も、推測だけで大規模修正せず観測事実を分離する。

### 2. Expected Behaviorを定義

「何が嫌だったか」だけでなく、同じ入力で本来どう振る舞うべきかを書く。

### 3. Root Cause分類

上記Diagnostic Categoryへ分類する。

例:

症状:
`feature-design` が毎回DB設計を人間へ聞く

可能な根本原因:
- feature-designのHuman Decisionルールが広すぎる
- decision-policyがData変更を一律L3にしている
- initial-designのdata ownershipが不足

### 4. Fix Locationを選択

最も上流かつ最小の正規ルールを直す。

- 共通問題 -> workflow policy
- 1 Skill固有 -> そのSKILL.md
- Generated Skillだけ -> そのGenerated Skillまたはbuilder
- Orchestration問題 -> planning-orchestrator

複数ファイルへ同じルールをコピペして直さない。

### 5. Minimal Patch

変更理由を説明できる最小差分にする。

禁止:

- 問題1件のためにSkill全体を書き直す
- 「絶対に質問しない」など過剰な対症療法
- 特定Feature名を汎用Skillへ埋め込む

### 6. Regression Scenario追加

`.ai/skill-tests/{target-skill}/` に再発防止シナリオを追加する。`scenario-template.yaml` を骨格としてよい。

Skillの発火条件・descriptionを変更した場合は、`trigger-eval-template.csv` と同じ形式で positive / negative trigger case も追加する。

最低限:

- id
- target_skill
- scenario
- input_summary
- expected
- forbidden
- related_issue

### 7. Regression Check

変更により逆方向の問題が出ないか確認する。

例:

「質問が多い」を直した結果、重要なArchitecture Decisionまで自走していないか。

少なくとも:

- 修正対象シナリオ
- 近接する正常シナリオ
- Human Decisionが必要な反対ケース

を確認する。

### 8. Final Structural Validation

修正後に再度 `validate_skillset.py` を実行し、構造上のregressionがないことを確認する。

### 9. Change Record

Skill修正の要約を残す。

推奨:

` .ai/skill-tests/CHANGELOG.md `

記録項目:

- 日付
- 対象Skill
- 症状
- Root Cause
- 修正
- Regression Scenario

## Human Decision Points

Skill MaintainerはSkill改善のためにProjectのProduct Decisionを変更してはならない。

以下は人間確認:

- Skill Architecture自体を大きく再編する
- Artifact ownershipを複数Skillに跨って変更する
- Planning/Implementation境界を変更する
- Decision PolicyのL3/L4基準を緩和し、安全性・責任範囲に大きく影響する

局所的な文言・手順・trigger修正は自走可能。

## Exit Criteria

- 問題のRoot Causeが明示されている
- 修正対象が正規のOwner/Policyに限定されている
- 最小差分で修正されている
- Regression Scenarioが追加されている
- 元の問題が改善し、重要なDecision Escalationを壊していない

## Next

修正したSkillを元の問題シナリオで再実行する。
