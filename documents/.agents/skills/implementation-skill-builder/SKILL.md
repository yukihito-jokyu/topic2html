---
name: implementation-skill-builder
description: Create or update project-specific implementation skills from approved planning handoffs, architecture/technology constraints, repository conventions, AGENTS.md, and existing code patterns. Use when implementation is ready but the project lacks reusable technology-specific skills, or existing implementation skills no longer match the stack. Do not implement the feature itself and do not encode new business or architectural decisions without approval.
---

# Implementation Skill Builder

## Goal

Planning領域から独立した、**プロジェクト・技術スタック固有のImplementation Skill**を必要な分だけ生成・更新する。

このSkill自身はFeatureを実装しない。
実装の再現性を高めるためのSkillを作る。

## Preconditions

少なくとも1つのFeatureに `implementation-handoff.yaml` が存在する。

また、可能な限り以下が利用可能であること。

- `design/**`
- `AGENTS.md`
- 既存コード
- lint / test / build設定
- 既存Implementation Skill registry

## Inputs

優先順:

1. `implementation-handoff.yaml`
2. approved Initial Design / Decisions
3. `AGENTS.md`
4. repository structure
5. existing production code conventions
6. build / test / lint configuration
7. existing generated implementation skills

## Owned Outputs

- `.agents/skills/impl-*/SKILL.md`
- 必要に応じて各Skillの `references/`, `scripts/`, `assets/`
- `.ai/workflow/implementation-skills.yaml`

生成時は `assets/implementation-skill-template.md` を骨格として使い、プロジェクト固有情報で具体化する。

Generated Skillは `.agents/skills/` の直接の子ディレクトリとして作ることを基本とする。

## Core Principle

**TaskごとにSkillを作らない。**

Skill化するのは、複数Task/Featureで再利用される実装判断の集合。

Skill化候補は次を満たすこと。

- 繰り返し利用される
- 一貫した実装判断が必要
- 技術またはArchitectureに依存する
- 単なる一般知識ではなくプロジェクト固有の制約・手順がある

例:

- `impl-go-domain`
- `impl-go-application`
- `impl-wails-binding`
- `impl-react-feature`
- `impl-db-repository`
- `impl-project-testing`

不適切な例:

- `impl-testconnection-timeout`
- `impl-add-button`
- `impl-string-validation`

## Autonomous Actions

- Tech Stack検出
- Repository conventions分析
- 既存コードから反復パターン抽出
- Implementation Handoffのlogical areaと必要Capabilityを対応付け
- 既存Skillで足りるか評価
- 新規Skill候補の統合・分割
- SKILL.md生成
- Skill registry更新
- 明白に古くなった生成Skillの更新提案

## Human Decision Points

以下はSkillに勝手に埋め込まない。

- 新しいArchitecture rule
- 新しいlibrary / framework採用
- Business Rule
- Security posture変更
- Repository全体のcoding convention変更

既存情報だけでは決まらない場合はL3/L4として人間へ上げる。

## Procedure

### 0. 実装Rootの運用入口を確認する

実装対象がPlanning成果物のあるRootと異なる場合、Implementation Skillを配置するRootに以下があるか確認する。

- Rootの`AGENTS.md`
- Planning workflowへ到達するorchestration entrypoint

欠けている場合、Generated Skillだけを置いて実装を開始可能と扱わない。`skill-maintainer`で最小のRoot運用入口を補い、Planning正本の場所を明示する。実装RootへPlanning成果物を複製しない。

### 1. Capability Inventory

現在のImplementation Handoff群から、繰り返し必要になる実装Capabilityを抽出する。

### 2. Existing Skill Match

`.ai/workflow/implementation-skills.yaml` と `.agents/skills/impl-*` を確認する。

各Capabilityについて:

- covered
- partially covered
- missing
- obsolete

を判定する。

### 3. Skill Boundary Design

Skillを技術名だけで分けず、**一貫した実装責務**で境界づける。

例:

`impl-go-backend` 1個が巨大すぎるなら、domain/application/repository等へ分割可能。
ただし小さくしすぎない。

### 4. Generated SKILL.md Requirements

生成する各Skillは最低限以下を持つ。

- YAML frontmatter: `name`, `description`
- Goal
- Trigger / Non-trigger scope
- Required inputs
- Project conventions
- Implementation procedure
- File placement rules（この段階では技術依存なので記述可能）
- Validation commands
- Allowed autonomous decisions
- Human / architecture escalation rules
- Completion criteria

### 5. Evidence-based Convention

規約は次の根拠優先順位で採用する。

1. AGENTS.md / 明示規約
2. Approved Design Decision
3. repo-wide dominant existing pattern
4. official technology convention

1〜3がない場合に「好み」をProject Ruleとして固定しない。

### 6. Registry更新

例:

```yaml
skills:
  - name: impl-go-domain
    path: .agents/skills/impl-go-domain/SKILL.md
    capabilities:
      - domain-model
      - domain-validation
    evidence:
      - AGENTS.md
      - design/architecture.md
    status: active
```

### 7. Dry Handoff Check

現在の `implementation-handoff.yaml` を各Generated Skillへ割り当てられるか確認する。

割り当て不能なTaskがある場合:

- Skill不足なのか
- TaskがFeature固有すぎてSkill不要なのか

を区別する。

Skill不要なTaskは通常のエージェント判断で実装してよい。

## Guardrails

- Product codeを実装しない
- Planning Artifactを書き換えない
- 1 Task = 1 Skillにしない
- 一般的な言語知識だけを長大に複製しない
- 既存コード1例だけからProject Ruleを断定しない
- Generated Skill内へFeature固有Requirementを恒久ルールとして埋め込まない

## Exit Criteria

- 現在のImplementation Handoffを実装するために必要な再利用可能CapabilityがSkillで十分カバーされている
- Generated SkillがProjectの実際の技術・規約に根拠を持つ
- Skill registryが更新されている
- 新しいArchitecture/Product Decisionを暗黙に導入していない

## Next

対応するGenerated Implementation Skillを使って実装する。

運用上の不具合が発生したら `skill-maintainer`。
