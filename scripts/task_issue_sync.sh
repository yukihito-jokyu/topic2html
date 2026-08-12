#!/usr/bin/env bash
set -euo pipefail

# Task Mapを検証し、GitHub Task Issueの生成内容・接続台帳を同じ規則から
# 作成する。Bashは制御、jqは区切り文字と改行を壊さないJSON処理、rtkは
# git・gh実行の記録と出力量制御に使う。必要な外部Commandはjq、rtk、git、gh。
# 対応版はBash 3.2以上、jq 1.6以上とする。
SCRIPT_DIR=${BASH_SOURCE[0]%/*}
ROOT=$(cd "$SCRIPT_DIR/.." && pwd -P)
TASK_MAP="$ROOT/docs/task-map.md"
REPO_DEFAULT="yukihito-jokyu/knowledge"
ROOT_ISSUE=1

abort_with() {
  printf 'ERROR: %s\n' "$1" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || abort_with "$1 が必要です（$2）"
}

require_supported_versions() {
  local jq_version jq_major jq_minor
  if ((BASH_VERSINFO[0] < 3 || (BASH_VERSINFO[0] == 3 && BASH_VERSINFO[1] < 2))); then
    abort_with 'Bash 3.2以上が必要です（配列・正規表現・process substitutionを同じ規則で実行するため）'
  fi
  jq_version=$(jq --version)
  [[ $jq_version =~ ^jq-([0-9]+)\.([0-9]+) ]] || abort_with "jqの版を判定できません: $jq_version"
  jq_major=${BASH_REMATCH[1]}
  jq_minor=${BASH_REMATCH[2]}
  if ((jq_major < 1 || (jq_major == 1 && jq_minor < 6))); then
    abort_with 'jq 1.6以上が必要です（scan・capture・JSON変換を同じ規則で実行するため）'
  fi
}

trim() {
  local value=$1
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

task_kind() {
  case $1 in
    L[0-9]) printf 'L' ;;
    *-S[0-9]*) printf 'S' ;;
    *) printf 'M' ;;
  esac
}

parent_id() {
  local id=$1
  case $(task_kind "$id") in
    L) printf '' ;;
    M) printf '%s' "${id%-M*}" ;;
    S) printf '%s' "${id%-S*}" ;;
  esac
}

parse_tasks() {
  local source=$1 line in_ledger=false rows='' empty id name plan_state execution_state deliverable dependency rest
  while IFS= read -r line || [[ -n $line ]]; do
    if [[ $in_ledger == false ]]; then
      [[ $line == '## タスク台帳' ]] && in_ledger=true
      continue
    fi
    [[ $line == '## '* ]] && break
    [[ $line == '| L'* ]] || continue
    [[ $line =~ \|[[:space:]]*$ ]] || continue
    IFS='|' read -r empty id name plan_state execution_state deliverable dependency rest <<<"$line"
    id=$(trim "${id:-}")
    [[ $id =~ ^L[1-6](-M[0-9]+)?(-S[0-9]+)?$ ]] || continue
    name=$(trim "${name:-}")
    plan_state=$(trim "${plan_state:-}")
    execution_state=$(trim "${execution_state:-}")
    deliverable=$(trim "${deliverable:-}")
    dependency=$(trim "${dependency:-}")
    rows+=$(jq -nc --arg id "$id" --arg name "$name" --arg plan_state "$plan_state" \
      --arg execution_state "$execution_state" --arg deliverable "$deliverable" --arg dependency "$dependency" \
      '{id:$id,name:$name,plan_state:$plan_state,execution_state:$execution_state,deliverable:$deliverable,dependency:$dependency}')$'\n'
  done <<<"$source"
  [[ $in_ledger == true ]] || abort_with 'タスク台帳が見つかりません'
  jq -sc '.' <<<"$rows"
}

validate_tasks() {
  local tasks=$1 duplicates kind expected actual bad missing id parent
  duplicates=$(jq -r 'group_by(.id)[] | select(length > 1) | .[0].id' <<<"$tasks")
  [[ -z $duplicates ]] || abort_with "Task ID重複: $(jq -Rrsc 'split("\n") | map(select(length>0)) | join(", ")' <<<"$duplicates")"
  for kind in L M S; do
    case $kind in L) expected=6 ;; M) expected=19 ;; S) expected=116 ;; esac
    actual=$(jq --arg kind "$kind" '[.[] | select((if (.id|test("-S[0-9]+$")) then "S" elif (.id|test("-M[0-9]+$")) then "M" else "L" end) == $kind)] | length' <<<"$tasks")
    [[ $actual -eq $expected ]] || abort_with "${kind}件数が${actual}件です（期待: ${expected}）"
  done
  bad=$(jq -r '.[] | select(.plan_state != "承認済み") | .id' <<<"$tasks")
  [[ -z $bad ]] || abort_with '承認済みでないTaskがあります'
  while IFS= read -r id; do
    [[ -n $id ]] || continue
    parent=$(parent_id "$id")
    jq -e --arg id "$parent" 'any(.[]; .id == $id)' >/dev/null <<<"$tasks" || abort_with "親Task欠落: $id -> $parent"
  done < <(jq -r '.[] | .id | select(test("-M|-S"))' <<<"$tasks")
}

read_tasks_from_file() {
  local source tasks
  source=$(<"$TASK_MAP")
  tasks=$(parse_tasks "$source")
  validate_tasks "$tasks"
  printf '%s' "$tasks"
}

task_field() {
  jq -r --arg id "$1" --arg field "$2" '.[] | select(.id == $id) | .[$field]' <<<"$TASKS_JSON"
}

task_exists() {
  jq -e --arg id "$1" 'any(.[]; .id == $id)' >/dev/null <<<"$TASKS_JSON"
}

uniq_lines() {
  jq -Rrsc 'split("\n") | map(select(length > 0)) | (reduce .[] as $value ([]; if index($value) then . else .+[$value] end))[]'
}

expand_leaf_refs() {
  local text=$1 l m first tail operator value current refs=''
  while IFS=$'\t' read -r l m first tail; do
    [[ -n ${l:-} ]] || continue
    refs+="L${l}-M${m}-S${first}"$'\n'
    current=$first
    while IFS=$'\t' read -r operator value; do
      [[ -n ${operator:-} ]] || continue
      if [[ $operator == '〜' ]]; then
        (( value >= current )) || abort_with "逆順のleaf範囲: $text"
        while (( ++current <= value )); do refs+="L${l}-M${m}-S${current}"$'\n'; done
      else
        refs+="L${l}-M${m}-S${value}"$'\n'
        current=$value
      fi
    done < <(jq -nr --arg tail "$tail" '$tail | scan("([・〜])S([0-9]+)") | @tsv')
  done < <(jq -nr --arg text "${text%%。出力を*}" '$text | scan("L([0-9]+)-M([0-9]+)-S([0-9]+)((?:[・〜]S[0-9]+)*)") | @tsv')
  printf '%s' "$refs" | uniq_lines
}

expand_parent_refs() {
  local text=$1 cleaned l first tail operator value current refs='' range_first range_last
  cleaned=$(jq -nr --arg text "$text" '$text | gsub("L[0-9]+-M[0-9]+-S[0-9]+(?:[・〜]S[0-9]+)*"; "")')
  while IFS=$'\t' read -r l first tail; do
    [[ -n ${l:-} ]] || continue
    refs+="L${l}-M${first}"$'\n'
    current=$first
    while IFS=$'\t' read -r operator value; do
      [[ -n ${operator:-} ]] || continue
      if [[ $operator == '〜' ]]; then
        while (( ++current <= value )); do refs+="L${l}-M${current}"$'\n'; done
      else
        refs+="L${l}-M${value}"$'\n'; current=$value
      fi
    done < <(jq -nr --arg tail "$tail" '$tail | scan("([・〜])M([0-9]+)") | @tsv')
  done < <(jq -nr --arg text "$cleaned" '$text | scan("L([0-9]+)-M([0-9]+)((?:[・〜]M[0-9]+)*)") | @tsv')
  while IFS=$'\t' read -r range_first range_last; do
    [[ -n ${range_first:-} ]] || continue
    while (( range_first <= range_last )); do refs+="L${range_first}"$'\n'; ((range_first += 1)); done
  done < <(jq -nr --arg text "$cleaned" '$text | scan("L([0-9]+)〜L([0-9]+)") | @tsv')
  printf '%s' "$refs" | uniq_lines
}

task_refs() {
  local text=$1 refs='' cleaned id
  refs+=$(expand_leaf_refs "$text")$'\n'
  refs+=$(expand_parent_refs "$text")$'\n'
  cleaned=$(jq -nr --arg text "$text" '$text | gsub("L[0-9]+-M[0-9]+-S[0-9]+(?:[・〜]S[0-9]+)*"; "")')
  while IFS= read -r id; do
    [[ -n $id ]] && task_exists "$id" && refs+="$id"$'\n'
  done < <(jq -nr --arg text "$cleaned" '$text | scan("L[0-9]+(?:-M[0-9]+)?")')
  while IFS= read -r id; do
    [[ -n $id ]] && task_exists "$id" && printf '%s\n' "$id"
  done < <(printf '%s' "$refs" | uniq_lines)
}

decision_for() {
  local id=$1 key
  [[ $(task_kind "$id") == S ]] && key=$(parent_id "$id") || key=$id
  case $key in
    L1-M1) printf 004 ;; L1-M2) printf 005 ;; L2-M1) printf 007 ;; L2-M2) printf 010 ;; L2-M3) printf 014 ;;
    L3-M1) printf 016 ;; L3-M2) printf 017 ;; L4-M1) printf 019 ;; L4-M2) printf 020 ;; L4-M3) printf 021 ;;
    L5-M1) printf 023 ;; L5-M2) printf 024 ;; L5-M3) printf 025 ;; L5-M4) printf 026 ;;
    L6-M1) printf 028 ;; L6-M2) printf 029 ;; L6-M3) printf 030 ;; L6-M4) printf 031 ;; L6-M5) printf 032 ;;
    L1) printf '001・002' ;; L2) printf 006 ;; L3) printf 015 ;; L4) printf 018 ;; L5) printf 022 ;; L6) printf '027・033' ;;
    *) abort_with "決定ID規則がありません: $id" ;;
  esac
}

task_type() {
  local id=$1 name=$2
  if [[ $id =~ ^L6-M[45]- ]]; then printf evaluation
  elif [[ $name =~ 実装|接続|実行Artifact|Black-box検証 ]]; then printf implementation
  else printf design; fi
}

path_for() {
  case $1 in
    L1-M1-S1) printf 'docs/design/knowledge-model/logical-schema.md' ;; L1-M1-S2) printf 'docs/design/knowledge-model/relations-and-integrity.md' ;;
    L1-M1-S3) printf 'docs/design/knowledge-model/evidence-derived-state.md' ;; L1-M1-S4) printf 'docs/design/knowledge-model/update-operations-and-lineage.md' ;;
    L1-M1-S5) printf 'docs/design/knowledge-model/requirements-traceability.md、docs/design/knowledge-model/README.md' ;;
    L1-M2-S1) printf 'docs/design/cli-contract/command-catalog.md' ;; L1-M2-S2) printf 'docs/design/cli-contract/collection-contract.md' ;;
    L1-M2-S3) printf 'docs/design/cli-contract/errors-and-exit-codes.md' ;; L1-M2-S4) printf 'docs/design/cli-contract/schemas/**' ;;
    L1-M2-S5) printf 'docs/design/cli-contract/versioning-and-compatibility.md' ;; L1-M2-S6) printf 'docs/design/cli-contract/requirements-traceability.md、docs/design/cli-contract/README.md' ;;
    L2-M1-S1) printf 'docs/design/knowledge-store/adr/0001-persistence-stack.md' ;; L2-M1-S2) printf 'docs/design/knowledge-store/physical-model-history-schema-evolution.md' ;;
    L2-M1-S3) printf 'docs/design/knowledge-store/repository-transaction-boundary.md、docs/design/knowledge-store/README.md' ;;
    L2-M2-S1) printf 'docs/design/search-infrastructure/search-and-index-requirements.md' ;; L2-M2-S2) printf 'docs/design/search-infrastructure/adr/0001-search-index-stack.md' ;;
    L2-M2-S3) printf 'docs/design/search-infrastructure/candidate-search-architecture.md' ;; L2-M2-S4) printf 'docs/design/search-infrastructure/index-lifecycle-recovery.md、docs/design/search-infrastructure/README.md' ;;
    L2-M3-S1) printf 'docs/design/cli-runtime/adr/0001-runtime-validation-distribution.md' ;; L2-M3-S2) printf 'docs/design/cli-runtime/command-to-port-architecture.md、docs/design/cli-runtime/README.md' ;;
    L3-M1-S1) printf 'docs/design/knowledge-acquisition/episode-input-evidence-boundary.md' ;; L3-M1-S2) printf 'docs/design/knowledge-acquisition/candidate-normalization-field-mapping.md' ;;
    L3-M1-S3) printf '.agents/skills/knowledge-acquisition/references/candidate-markdown-contract.md' ;; L3-M1-S4) printf '.agents/skills/knowledge-acquisition/SKILL.md、.agents/skills/knowledge-acquisition/references/**（S3所有Fileを除く）' ;;
    L3-M2-S1) printf 'docs/design/knowledge-update/candidate-acceptance.md' ;; L3-M2-S2) printf 'docs/design/knowledge-update/existing-knowledge-comparison.md' ;;
    L3-M2-S3) printf '.agents/skills/knowledge-update/references/update-decision-contract.md' ;; L3-M2-S4) printf '.agents/skills/knowledge-update/references/cli-command-mapping.md' ;;
    L3-M2-S5) printf '.agents/skills/knowledge-update/SKILL.md、.agents/skills/knowledge-update/references/**（S3・S4所有Fileを除く）' ;; L3-M2-S6) printf 'tests/component/knowledge-update/**' ;;
    L4-M1-S1) printf 'docs/design/article-analysis/article-input-boundary.md' ;; L4-M1-S2) printf 'docs/design/article-analysis/overview-claim-decomposition.md' ;;
    L4-M1-S3) printf 'docs/design/article-analysis/location-support-traceability.md' ;; L4-M1-S4) printf '.agents/skills/article-analysis/references/article-analysis-markdown-contract.md' ;;
    L4-M1-S5) printf '.agents/skills/article-analysis/SKILL.md、.agents/skills/article-analysis/references/**（S4所有Fileを除く）' ;;
    L4-M2-S1) printf 'docs/design/knowledge-search/agentic-search-requirements.md' ;; L4-M2-S2) printf 'docs/design/knowledge-search/target-claim-query-reconstruction.md' ;;
    L4-M2-S3) printf '.agents/skills/knowledge-search/references/cli-command-mapping.md' ;; L4-M2-S4) printf 'docs/design/knowledge-search/assessment-decision.md' ;;
    L4-M2-S5) printf '.agents/skills/knowledge-search/references/knowledge-search-output-contract.md' ;; L4-M2-S6) printf '.agents/skills/knowledge-search/SKILL.md、.agents/skills/knowledge-search/references/**（S3・S5所有Fileを除く）' ;;
    L4-M2-S7) printf 'tests/component/knowledge-search/**' ;;
    L4-M3-S1) printf 'docs/design/reading-value/input-alignment-and-evaluability.md' ;; L4-M3-S2) printf 'docs/design/reading-value/recognition-gain-application.md' ;;
    L4-M3-S3) printf 'docs/design/reading-value/reliability-applicability.md' ;; L4-M3-S4) printf 'docs/design/reading-value/recommendation-range-aggregation.md' ;;
    L4-M3-S5) printf '.agents/skills/reading-value/references/reading-value-output-contract.md' ;; L4-M3-S6) printf '.agents/skills/reading-value/SKILL.md、.agents/skills/reading-value/references/**（S5所有Fileを除く）' ;;
    L5-M1-S1) printf 'docs/design/orchestration/common-control-contract.md' ;; L5-M1-S2) printf 'docs/design/orchestration/child-skill-registry-contract.md' ;;
    L5-M1-S3) printf 'docs/design/orchestration/execution-state-envelope.md' ;; L5-M1-S4) printf 'docs/design/orchestration/correlation-trace-contract.md' ;; L5-M1-S5) printf 'docs/design/orchestration/execution-control-policy.md' ;;
    L5-M2-S1) printf 'docs/design/orchestration/article-reading-workflow.md' ;; L5-M2-S2) printf 'docs/design/orchestration/article-reading/article-search-fanout.md' ;;
    L5-M2-S3) printf 'docs/design/orchestration/article-reading/article-search-fanin-handoff.md' ;; L5-M2-S4) printf 'docs/design/orchestration/article-reading/article-research-cycle-routing.md' ;; L5-M2-S5) printf 'docs/design/orchestration/article-reading/article-workflow-termination.md' ;;
    L5-M3-S1) printf 'docs/design/orchestration/knowledge-accumulation-workflow.md' ;; L5-M3-S2) printf 'docs/design/orchestration/knowledge-accumulation/acquisition-result-branching.md' ;;
    L5-M3-S3) printf 'docs/design/orchestration/knowledge-accumulation/update-work-items.md' ;; L5-M3-S4) printf 'docs/design/orchestration/knowledge-accumulation/update-result-resumption.md' ;; L5-M3-S5) printf 'docs/design/orchestration/knowledge-accumulation/workflow-termination.md' ;;
    L5-M4-S1) printf '.agents/skills/parent-orchestration/SKILL.md、.agents/skills/parent-orchestration/references/common/**、.agents/skills/parent-orchestration/references/child-registry.md' ;;
    L5-M4-S2) printf '.agents/skills/parent-orchestration/references/workflows/article-reading.md' ;; L5-M4-S3) printf '.agents/skills/parent-orchestration/references/workflows/knowledge-accumulation.md' ;;
    L5-M4-S4) printf 'tests/contract/parent-orchestration/**' ;; L5-M4-S5) printf 'tests/component/parent-orchestration/knowledge-accumulation/**' ;; L5-M4-S6) printf 'tests/component/parent-orchestration/article-reading/**' ;;
    L6-M1-S1) printf 'docs/design/evaluation/spec/evaluation-layer-policy.md' ;; L6-M1-S2) printf 'docs/design/evaluation/spec/coverage-responsibility-matrix.md' ;;
    L6-M1-S3) printf 'docs/design/evaluation/spec/verdict-criteria.md' ;; L6-M1-S4) printf 'docs/design/evaluation/spec/search-trace-diagnostics.md' ;; L6-M1-S5) printf 'docs/design/evaluation/spec/evaluation-report-contract.md' ;;
    L6-M2-S1) printf 'docs/design/evaluation/datasets/dataset-contract.md' ;; L6-M2-S2) printf 'docs/design/evaluation/datasets/scenario-catalog.md' ;;
    L6-M2-S3) printf 'tests/evaluation/datasets/schema/**、tests/evaluation/datasets/tools/**' ;; L6-M2-S4) printf 'tests/evaluation/datasets/scenarios/a-e/**' ;;
    L6-M2-S5) printf 'tests/evaluation/datasets/scenarios/f-h/**' ;; L6-M2-S6) printf 'tests/evaluation/datasets/scenarios/i-j-reading-value/**' ;;
    L6-M3-S1) printf 'docs/design/evaluation/harness/harness-requirements.md' ;; L6-M3-S2) printf 'docs/design/evaluation/adr/evaluation-harness-stack.md' ;;
    L6-M3-S3) printf 'docs/design/evaluation/harness/harness-architecture.md' ;; L6-M3-S4) printf 'tests/evaluation/harness/core/**' ;;
    L6-M3-S5) printf 'tests/evaluation/harness/adapters/cli/**' ;; L6-M3-S6) printf 'tests/evaluation/harness/adapters/skills/**' ;; L6-M3-S7) printf 'tests/evaluation/harness/adapters/workflows/**' ;;
    L6-M5-S1) printf 'tests/evaluation/workflows/knowledge-accumulation/**、tests/evaluation/reports/workflows/knowledge-accumulation/**' ;;
    L6-M5-S2) printf 'tests/evaluation/workflows/article-reading/**、tests/evaluation/reports/workflows/article-reading/**' ;;
    L6-M5-S3) printf 'tests/evaluation/workflows/regression/**、tests/evaluation/reports/final/**' ;;
    L2-*) printf 'L2 Design Freeze Gate通過記録の%s専用Path（Issue作成時TBD、Ready前に実値化）' "$1" ;;
    L6-M4-S1) printf 'tests/evaluation/suites/cli/**、tests/evaluation/reports/cli/**' ;;
    L6-M4-S2) printf 'tests/evaluation/suites/l3/knowledge-acquisition/**、tests/evaluation/reports/l3/knowledge-acquisition/**' ;;
    L6-M4-S3) printf 'tests/evaluation/suites/l3/knowledge-update/**、tests/evaluation/reports/l3/knowledge-update/**' ;;
    L6-M4-S4) printf 'tests/evaluation/suites/l4/article-analysis/**、tests/evaluation/reports/l4/article-analysis/**' ;;
    L6-M4-S5) printf 'tests/evaluation/suites/l4/knowledge-search/**、tests/evaluation/reports/l4/knowledge-search/**' ;;
    L6-M4-S6) printf 'tests/evaluation/suites/l4/reading-value/**、tests/evaluation/reports/l4/reading-value/**' ;;
    *) abort_with "Path規則がありません: $1" ;;
  esac
}

validate_paths() {
  local paths='' id path collisions
  while IFS= read -r id; do path=$(path_for "$id"); paths+=$(jq -nc --arg id "$id" --arg path "$path" '{id:$id,path:$path}')$'\n'; done < <(jq -r '.[]|select(.id|test("-S[0-9]+$"))|.id' <<<"$TASKS_JSON")
  collisions=$(jq -Rsc 'split("\n")|map(select(length>0)|fromjson)|group_by(.path)[]|select(length>1)|"\(.[0].path)=\(map(.id)|join(","))"' <<<"$paths")
  [[ -z $collisions ]] || abort_with "leaf間で成果物Pathが重複しています: $(jq -Rsc 'split("\n")|map(select(length>0))|join("; ")' <<<"$collisions")"
}

mapping_get() { jq -r --arg id "$1" '.[$id] // empty' <<<"$MAPPING_JSON"; }
issue_link() { printf '#%s `[%s] %s`' "$1" "$2" "$(task_field "$2" name)"; }

linked_dependencies() {
  local id=$1 dependency refs ref number rendered=''
  dependency=$(task_field "$id" dependency); refs=$(task_refs "$dependency")
  [[ -n $refs ]] || { printf 'なし'; return; }
  while IFS= read -r ref; do
    number=$(mapping_get "$ref")
    if [[ -n $number ]]; then rendered+="$(issue_link "$number" "$ref")、"; else rendered+="\`$ref\`、"; fi
  done <<<"$refs"
  printf '%s' "${rendered%、}"
}

dependency_kind() {
  local label=$1 clause=$2
  case $label in 完了|完了・Merge|Report確定|確定|ADR確定) printf '完了・Merge依存'; return ;; Gateへの入力) printf 'Gateへの入力'; return ;; 完了・Release|Release条件) printf 'Release条件'; return ;; 本番実装|Mock実装|本番Skill実装) printf 'Gate通過依存'; return ;; esac
  [[ $clause == *Releaseは* ]] && printf 'Release条件' || printf '着手依存'
}

dependency_rows() {
  local id=$1 dependency clause label value kind refs ref number rendered milestones grouped=''
  dependency=$(task_field "$id" dependency)
  while IFS= read -r clause; do
    clause=$(trim "$clause"); [[ -n $clause && $clause != 出力を* ]] || continue
    label=''; value=$clause
    if [[ $clause =~ ^([^:：]+)[:：][[:space:]]*(.*)$ ]]; then label=${BASH_REMATCH[1]}; value=${BASH_REMATCH[2]}; fi
    kind=$(dependency_kind "$label" "$clause"); rendered=''; refs=$(task_refs "$value")
    while IFS= read -r ref; do
      [[ -n $ref ]] || continue; number=$(mapping_get "$ref")
      [[ -n $number ]] && rendered+="$(issue_link "$number" "$ref")"$'\n' || rendered+="\`$ref\`（Issue作成待ち）"$'\n'
    done <<<"$refs"
    milestones=$(jq -nr --arg text "$value" '$text|scan("(?:L[0-9]+(?:-M[0-9]+)?\\s+)?(?:Design Freeze Gate|Target Readiness Gate|Final Quality Gate)|L[0-9]+ Release")')
    while IFS= read -r ref; do [[ -n $ref ]] && rendered+="\`$ref\`"$'\n'; done <<<"$milestones"
    [[ -n $rendered ]] || rendered="$value"$'\n'
    while IFS= read -r ref; do [[ -n $ref ]] && grouped+=$(jq -nc --arg kind "$kind" --arg value "$ref" '{kind:$kind,value:$value}')$'\n'; done <<<"$(printf '%s' "$rendered" | uniq_lines)"
  done < <(jq -nr --arg text "$dependency" '$text|split("。")[]')
  for kind in '着手依存' '完了・Merge依存' 'Gateへの入力' 'Gate通過依存' 'Release条件'; do
    rendered=$(jq -Rrsc --arg kind "$kind" 'split("\n")|map(select(length>0)|fromjson)|map(select(.kind==$kind).value)|reduce .[] as $value ([]; if index($value) then . else .+[$value] end)|join("、")' <<<"$grouped")
    [[ -n $rendered ]] && printf '| %s | %s | Task Map原文の依存時点を満たす |\n' "$kind" "$rendered" || printf '| %s | 該当なし | Task Mapに直接条件なし |\n' "$kind"
  done
}

human_progress() {
  local kind=$1 text
  if [[ $kind == S ]]; then
    IFS= read -r -d '' text <<'EOF' || true

<!-- human-progress:start -->
## 実行記録（手動更新欄）

### 実施チェック

- [ ] 主成果物を作成した
- [ ] 直接依存とGate条件を確認した
- [ ] Owner／Path境界を守った
- [ ] Task固有の検証を完了した
- [ ] 未解決の契約差異とReady後のTBDがない

### 着手・検証Evidence

- worktree起点SHA:
- Branch:
- 依存TaskのMerge commit:
- Gate通過記録:
- 検証Command／Review:
- 結果:
- 関連PR:
<!-- human-progress:end -->
EOF
  else
    IFS= read -r -d '' text <<'EOF' || true

<!-- human-progress:start -->
## 実行記録（手動更新欄）

- [ ] 直下の子Issueがすべて完了している
- [ ] 親の到達状態が子成果物の総和として成立している
- [ ] Gate／Release条件を満たしている

- 統合Commit／Release evidence:
- 未解決事項:
<!-- human-progress:end -->
EOF
  fi
  printf '%s' "$text"
}

generated_wrap() {
  local content=${2%$'\n'}
  printf '<!-- knowledge-task-id: %s -->\n<!-- generated-content:start -->\n%s\n<!-- generated-content:end -->\n' "$1" "$content"
}

state_for() {
  local states='{}'
  [[ -n ${ISSUE_STATES_JSON:-} ]] && states=$ISSUE_STATES_JSON
  jq -r --arg id "$1" '.[$id] // empty' <<<"$states"
}

tracking_body() {
  local id=$1 kind parent parent_number child_pattern children='' child number checked content
  kind=$(task_kind "$id")
  if [[ $kind == L ]]; then parent="原典Issue #$ROOT_ISSUE"; else parent_number=$(mapping_get "$(parent_id "$id")"); parent=$(issue_link "$parent_number" "$(parent_id "$id")"); fi
  [[ $kind == L ]] && child_pattern="^${id}-M[0-9]+$" || child_pattern="^${id}-S[0-9]+$"
  while IFS= read -r child; do
    number=$(mapping_get "$child"); [[ $(state_for "$child") == CLOSED ]] && checked=x || checked=' '
    [[ -n $number ]] && children+="- [$checked] $(issue_link "$number" "$child")"$'\n' || children+="- [ ] \`[$child] $(task_field "$child" name)\`（Issue作成待ち）"$'\n'
  done < <(jq -r --arg pattern "$child_pattern" '.[]|select(.id|test($pattern))|.id' <<<"$TASKS_JSON")
  IFS= read -r -d '' content <<EOF || true
## タスク情報

- Task ID: \`$id\`
- 親Task: $parent
- 区分: $([[ $kind == L ]] && printf '大分類tracking' || printf '中分類tracking')
- 原典Issue: #$ROOT_ISSUE
- 関連する決定ID: 決定$(decision_for "$id")

## 目的・主成果物

$(task_field "$id" name)。到達状態は「$(task_field "$id" deliverable)」。

## 直接依存

- Task Map原文: $(task_field "$id" dependency)
- 参照Task: $(linked_dependencies "$id")

## 直下の子Issue

${children%$'\n'}

## 完了条件

- 直下の子Issueがすべて完了している
- \`$(task_field "$id" deliverable)\`が子成果物の総和として成立している
- Task Mapに記載されたGate／Release条件を満たしている
- 未解決の契約差異がない

## 境界

このIssueは進捗と統合条件を追跡する。子Issueと重複する設計・実装・評価成果物を独自に作成しない。Gate／ReleaseはMilestoneであり、別Issueを作成しない。
EOF
  generated_wrap "$id" "$content"; human_progress "$kind"
}

leaf_body() {
  local id=$1 parent parent_number path name deliverable decision type rows content
  parent=$(parent_id "$id"); parent_number=$(mapping_get "$parent"); path=$(path_for "$id"); name=$(task_field "$id" name); deliverable=$(task_field "$id" deliverable); decision=$(decision_for "$id"); type=$(task_type "$id" "$name"); rows=$(dependency_rows "$id")
  IFS= read -r -d '' content <<EOF || true
## タスク情報

- Task ID: \`$id\`
- 親Task ID: \`$parent\`
- 親Issue: $(issue_link "$parent_number" "$parent")
- タスク名: $name
- タスク種別: \`$type\`
- 原典Issue: #$ROOT_ISSUE
- 関連する原典章: Issue #1、および決定${decision}に記録された対応章
- 関連する決定ID: 決定${decision}・034〜037

## 目的

${name}ことで、\`${deliverable}\`を完成させる。

## 原典との差分

- 入力: Issue #1、現在のTask Map、決定${decision}、直接依存の承認済み成果物
- 原典で決定済みだが未実施の成果物: $deliverable
- このleafで決める未確定事項: Task名が要求する詳細化または実装に限定する
- 選び直さない事項: Issue #1とTask Mapで確定した責務境界、依存DAG、Gate／Release、Codex／CLI境界

## 実施内容

- [ ] \`${deliverable}\`を指定Owner境界で作成する
- [ ] 直接依存とGate条件を確認し、Task固有のReviewまたは検証を記録する

## 成果物と所有

| 項目 | 内容 |
| --- | --- |
| 主成果物 | $deliverable |
| 書込み可能なPath／Glob | \`$path\` |
| 単一Owner | \`$id\` |
| read-only入力 | $(linked_dependencies "$id") |
| 共有資産と単一Owner | Task Mapのworktree所有境界に従う。共有物は記載Ownerへ変更を戻す |
| Gate通過記録 | Task Map原文のGate条件に従う。Gate非対象なら該当なし |

書込み可能なのは上表の所有Pathと承認済みの明示例外だけとし、それ以外は変更しない。

## 完了条件

- [ ] \`${deliverable}\`が作成され、Task名の到達状態を満たす
- [ ] Task Mapの直接依存、Gate、Releaseの必要条件を満たす
- [ ] 単一Ownerと書込みPathの境界を守る
- [ ] Task固有の検証結果またはReview evidenceをIssueへ記録する
- [ ] 未解決の契約差異と、Ready後のTBDがない

## 対象外

- Issue #1で既に決定した原則・責務の選び直し
- 別Taskが単一Ownerである成果物、共有評価、推移的依存の再実装

## 依存関係

| 種別 | Task／Milestone | 必要な成果物・状態 |
| --- | --- | --- |
${rows%$'\n'}

## 着手判定

| 確認項目 | 結果・Evidence |
| --- | --- |
| 現在のTask MapとIssueの依存関係・Gate・成果物Ownerが一致する | 着手時に確認 |
| 着手依存TaskのMerge commit | 着手時に記録 |
| 必要なGateの通過記録、またはGate前Taskであること | 着手時に記録 |
| worktree起点SHA | 着手時に記録 |
| 必須値に未解決TBDがない | 着手時に確認 |
| 並行Taskと書込みPathが競合しない | 着手時に確認 |

- [ ] 上表を確認し、このTaskは着手可能である

## worktree・Merge

| 項目 | 内容 |
| --- | --- |
| worktree起点SHA | 着手時に確定 |
| Branch | 着手時に確定 |
| 所有Path／Glob | \`$path\` |
| 共有物と単一Owner | Task Mapのworktree所有境界に従う |
| 並行可能Task | Pathが交差せず、必要な依存を満たすTask |
| 直列化するTaskと理由 | Task Mapの直接依存・Merge順に従う |
| Merge前提 | 直接依存の必要状態、Gate通過、Task固有検証合格 |
| Merge順 | Task Mapの承認済みDAGに従う |
| 統合先 | Task Mapの依存順に従う統合branch |

## 検証

| 種別 | 方法・Command | 合格条件 | Evidence |
| --- | --- | --- | --- |
| 静的確認 | 成果物・参照・Owner境界をReview | 不足・重複・境界違反がない | 完了時に記録 |
| Task固有テスト／Review | 成果物種別に応じた固有検証 | Task固有の完了条件を満たす | 完了時に記録 |
| 契約・統合確認 | Task Mapに明記された場合のみ実施 | 対象契約に適合する | 完了時に記録 |
| 後続評価 | L6等の後続Issueで実施 | このIssueでは重複実施しない | 後続Issueへ記録 |

## 差異を発見した場合

- [ ] 作業を停止する
- [ ] Issue内で新しい依存、Path、設計判断を決めない
- [ ] Task MapまたはGate記録の修正案を議論記録へ残す
- [ ] 必要な再承認後、Task MapとIssueを同期する

## 関連Issue・PR

- 原典Issue: #$ROOT_ISSUE
- 親Issue: $(issue_link "$parent_number" "$parent")
- Blocked by: $(linked_dependencies "$id")
- 関連PR: 未作成
EOF
  content=$(jq -nr --arg text "$content" '$text|gsub("- \\[ \\]";"- 要件:")')
  generated_wrap "$id" "$content"; human_progress S
}

body_for() { [[ $(task_kind "$1") == S ]] && leaf_body "$1" || tracking_body "$1"; }

gh_json() { rtk gh issue list --repo "$REPO" --state all --limit 500 --json number,title,body,url,state || abort_with 'gh issue list失敗'; }

managed_issues() {
  local issues=$1 duplicates
  duplicates=$(jq -r '[.[]|. + {task_id:(.body | try capture("<!-- knowledge-task-id: (?<id>L[0-9]+(?:-M[0-9]+)?(?:-S[0-9]+)?) -->").id catch null)}|select(.task_id)]|group_by(.task_id)[]|select(length>1)|.[0].task_id' <<<"$issues")
  [[ -z $duplicates ]] || abort_with "GitHub上のTask ID重複: $duplicates"
  jq '[.[]|. + {task_id:(.body | try capture("<!-- knowledge-task-id: (?<id>L[0-9]+(?:-M[0-9]+)?(?:-S[0-9]+)?) -->").id catch null)}|select(.task_id)]|map({key:.task_id,value:.})|from_entries' <<<"$issues"
}

title_conflicts() {
  local id count
  while IFS= read -r id; do
    count=$(jq --arg id "$id" '[.[]
      | select(.title|startswith("["+$id+"]"))
      | . + {marker_task_id:(.body | try capture("<!-- knowledge-task-id: (?<id>L[0-9]+(?:-M[0-9]+)?(?:-S[0-9]+)?) -->").id catch null)}
      | select(.marker_task_id != $id)] | length' <<<"$ISSUES_JSON")
    [[ $count -eq 0 ]] || abort_with "Markerなし／別Markerのタイトル競合: $id"
  done < <(jq -r '.[].id' <<<"$TASKS_JSON")
}

run_gh_with_body() {
  local body=$1
  shift
  printf '%s' "$body" | rtk gh "$@" || abort_with "gh $* 失敗"
}

json_text_field() {
  local destination=$1 json=$2 filter=$3 marker=$'\036' value
  value=$(jq -rj "$filter" <<<"$json"; printf '%s' "$marker")
  value=${value%"$marker"}
  printf -v "$destination" '%s' "$value"
}

create_issue() {
  local id=$1 body=$2 output number
  output=$(run_gh_with_body "$body" issue create --repo "$REPO" --title "[$id] $(task_field "$id" name)" --body-file -)
  number=${output##*/}; [[ $number =~ ^[0-9]+$ ]] || abort_with "Issue番号を取得できません: $output"
  CREATED_NUMBER=$number
  printf 'CREATE %s -> #%s\n' "$id" "$number"
}

replace_generated() {
  local existing=$1 desired=$2 destination=$3 starts ends desired_generated marker=$'\036' value
  starts=$(jq -nr --arg text "$existing" '$text|[scan("<!-- generated-content:start -->")]|length'); ends=$(jq -nr --arg text "$existing" '$text|[scan("<!-- generated-content:end -->")]|length')
  [[ $starts -eq 1 && $ends -eq 1 ]] || abort_with '管理Marker不整合のため本文を更新できません'
  desired_generated=$(jq -nr --arg text "$desired" '$text|capture("(?<part><!-- generated-content:start -->.*<!-- generated-content:end -->)";"m").part')
  value=$(jq -nrj --arg text "$existing" --arg replacement "$desired_generated" '$text|sub("<!-- generated-content:start -->.*<!-- generated-content:end -->";$replacement;"m")'; printf '%s' "$marker")
  value=${value%"$marker"}
  printf -v "$destination" '%s' "$value"
}

update_issue() {
  local id=$1 issue=$2 desired=$3 title existing merged number
  title="[$id] $(task_field "$id" name)"
  json_text_field existing "$issue" '.body // ""'
  replace_generated "$existing" "$desired" merged
  number=$(jq -r '.number' <<<"$issue")
  [[ $(jq -r '.title' <<<"$issue") == "$title" && $existing == "$merged" ]] && return
  run_gh_with_body "$merged" issue edit "$number" --repo "$REPO" --title "$title" --body-file - >/dev/null
  printf 'UPDATE %s -> #%s\n' "$id" "$number"
}

root_children_block() {
  local lines='' id number checked
  while IFS= read -r id; do number=$(mapping_get "$id"); [[ $(state_for "$id") == CLOSED ]] && checked=x || checked=' '; lines+="- [$checked] $(issue_link "$number" "$id")"$'\n'; done < <(jq -r '.[]|select(.id|test("^L[0-9]+$"))|.id' <<<"$TASKS_JSON")
  printf '<!-- task-map-children:start -->\n## Task Map: 大分類Issue\n\n現在のTask Mapに基づく直下の大分類tracking Issueです。Gate／ReleaseはIssue化していません。\n\n%s<!-- task-map-children:end -->' "$lines"
}

update_root_issue() {
  local issue=$1 body block desired number marker=$'\036' value
  json_text_field body "$issue" '.body // ""'
  block=$(root_children_block); number=$(jq -r '.number' <<<"$issue")
  if [[ $body == *'<!-- task-map-children:start -->'* && $body == *'<!-- task-map-children:end -->'* ]]; then
    value=$(jq -nrj --arg text "$body" --arg block "$block" '$text|sub("<!-- task-map-children:start -->.*<!-- task-map-children:end -->";$block;"m")'; printf '%s' "$marker")
    desired=${value%"$marker"}
  else
    desired="${body%$'\n'}"$'\n\n'"$block"$'\n'
  fi
  [[ $desired == "$body" ]] && return
  run_gh_with_body "$desired" issue edit "$number" --repo "$REPO" --body-file - >/dev/null
  printf 'UPDATE ROOT -> #%s\n' "$number"
}

verify_all() {
  local managed=$1 extra missing counts id issue expected_title body heading child_pattern child_section child count line_count root root_section
  extra=$(jq -r --argjson tasks "$TASKS_JSON" 'keys[] as $id|select(any($tasks[];.id==$id)|not)|$id' <<<"$managed"); [[ -z $extra ]] || abort_with "余分な管理Issue: $extra"
  missing=$(jq -r --argjson managed "$managed" '.[]|select($managed[.id]==null)|.id' <<<"$TASKS_JSON"); [[ -z $missing ]] || abort_with "不足する管理Issue: $missing"
  counts=$(jq '{L:([keys[]|select(test("^L[0-9]+$"))]|length),M:([keys[]|select(test("-M[0-9]+$"))]|length),S:([keys[]|select(test("-S[0-9]+$"))]|length)}' <<<"$managed")
  [[ $(jq -r '.L' <<<"$counts") -eq 6 && $(jq -r '.M' <<<"$counts") -eq 19 && $(jq -r '.S' <<<"$counts") -eq 116 ]] || abort_with "GitHub件数不一致: $counts"
  while IFS= read -r id; do
    issue=$(jq --arg id "$id" '.[$id]' <<<"$managed"); body=$(jq -r '.body // ""' <<<"$issue"); expected_title="[$id] $(task_field "$id" name)"
    [[ $(jq -r '.title' <<<"$issue") == "$expected_title" ]] || abort_with "タイトル不一致: $id"
    if [[ $(task_kind "$id") == S ]]; then
      for heading in タスク情報 目的 原典との差分 実施内容 成果物と所有 完了条件 対象外 依存関係 着手判定 worktree・Merge 検証 差異を発見した場合 関連Issue・PR; do [[ $body == *"## $heading"* ]] || abort_with "leaf必須Section欠落: $id $heading"; done
    else
      child_section=$(jq -nr --arg text "$body" '$text|capture("## 直下の子Issue\\n\\n(?<part>.*?)\\n\\n## 完了条件";"m").part // ""')
      [[ $(task_kind "$id") == L ]] && child_pattern="^${id}-M[0-9]+$" || child_pattern="^${id}-S[0-9]+$"
      while IFS= read -r child; do count=$(jq -nr --arg text "$child_section" --arg needle "[$child]" '$text|[indices($needle)[]]|length'); [[ $count -eq 1 ]] || abort_with "親子チェックリスト不一致: $id -> $child count=$count"; done < <(jq -r --arg pattern "$child_pattern" '.[]|select(.id|test($pattern))|.id' <<<"$TASKS_JSON")
      line_count=$(jq -nr --arg text "$child_section" '$text|split("\n")|map(select(test("^- \\[[ x]\\] #[0-9]+ ")))|length')
      count=$(jq -r --arg pattern "$child_pattern" '[.[]|select(.id|test($pattern))]|length' <<<"$TASKS_JSON")
      [[ $line_count -eq $count ]] || abort_with "親子チェックリスト件数不一致: $id"
    fi
  done < <(jq -r '.[].id' <<<"$TASKS_JSON")
  root=$(jq --argjson number "$ROOT_ISSUE" '.[]|select(.number==$number)' <<<"$ISSUES_JSON"); [[ -n $root ]] || abort_with "原典Issue #$ROOT_ISSUEがありません"
  root_section=$(jq -r '.body // ""' <<<"$root"); root_section=$(jq -nr --arg text "$root_section" '$text|capture("<!-- task-map-children:start -->(?<part>.*?)<!-- task-map-children:end -->";"m").part // ""')
  while IFS= read -r id; do count=$(jq -nr --arg text "$root_section" --arg needle "[$id]" '$text|[indices($needle)[]]|length'); [[ $count -eq 1 ]] || abort_with "原典Issueの大分類接続不一致: $id"; done < <(jq -r '.[]|select(.id|test("^L[0-9]+$"))|.id' <<<"$TASKS_JSON")
  printf 'VERIFY OK L=6 M=19 S=116 TOTAL=141\n'
}

render_connections() {
  local edges='' milestones='' consumer dependency clause label value phase raw producer gate milestone producer_deliverable id producers terminals
  while IFS= read -r consumer; do
    dependency=$(task_field "$consumer" dependency)
    while IFS= read -r clause; do
      clause=$(trim "$clause"); [[ -n $clause && $clause != 出力を* ]] || continue
      label=''; value=$clause; if [[ $clause =~ ^([^:：]+)[:：][[:space:]]*(.*)$ ]]; then label=${BASH_REMATCH[1]}; value=${BASH_REMATCH[2]}; fi
      case $label in 完了・Merge) phase='完了・Merge' ;; Report確定) phase='Report確定' ;; 確定|ADR確定) phase='確定' ;; Gateへの入力) phase='Gate入力' ;; *) [[ $clause == Mergeは* ]] && phase='完了・Merge' || { [[ $clause == *Releaseは* ]] && phase='Release条件' || phase='着手'; } ;; esac
      while IFS= read -r producer; do
        [[ -n $producer ]] || continue
        task_exists "$producer" || abort_with "接続先に存在しないleaf: $producer -> $consumer"
        [[ $(task_kind "$producer") == S ]] || abort_with "接続先に存在しないleaf: $producer -> $consumer"
        [[ $clause == *Gate* || $clause == *Release* ]] && gate=$clause || gate='—'; producer_deliverable=$(task_field "$producer" deliverable)
        edges+=$(jq -nc --arg p "$producer" --arg c "$consumer" --arg d "$producer_deliverable" --arg phase "$phase" --arg gate "$gate" --arg owner "Task Map「$(parent_id "$producer")内」" '[ $p,$c,$d,$phase,"read-only",$gate,$owner ]')$'\n'
      done < <(expand_leaf_refs "$value")
      if [[ $clause == *Gate* || $clause == *Release* ]]; then
        milestone=$(jq -nr --arg text "$clause" '$text|gsub("L[0-9]+-M[0-9]+-S[0-9]+(?:[・〜]S[0-9]+)*[、]?";"")|gsub("^[[:space:]]+|[[:space:]]+$";"")')
        [[ -n $milestone ]] && milestones+=$(jq -nc --arg m "$milestone" --arg c "$consumer" --arg p "$phase" '[$m,$c,$p]')$'\n'
      fi
    done < <(jq -nr --arg text "$dependency" '$text|split("。")[]')
  done < <(jq -r '.[]|select(.id|test("-S[0-9]+$"))|.id' <<<"$TASKS_JSON")
  edges=$(jq -Rsc --arg d "$(task_field L3-M2-S2 deliverable)" 'split("\n")|map(select(length>0)|fromjson)|map(select(.[0]!="L3-M2-S2" or .[1]!="L2-M2-S1"))+[["L3-M2-S2","L2-M2-S1",$d,"Gate入力","read-only","L2 Design Freeze Gate入力","Task Map「L3-M2内」"]]|unique|sort_by(.[0],.[1],.[3])' <<<"$edges")
  printf '| Producer leaf | 直接の後続利用者 | 利用成果物 | 依存時点 | 利用方法 | Gate／Release | worktree所有記録への参照 |\n| --- | --- | --- | --- | --- | --- | --- |\n'
  jq -r '.[]|"| "+join(" | ")+" |"' <<<"$edges"
  printf '\n### 終端leaf\n\n| Leaf Task | 後続利用者 | 主成果物 |\n| --- | --- | --- |\n'
  producers=$(jq -c '[.[] | .[0]] | unique' <<<"$edges")
  jq -r --argjson producers "$producers" '.[]|select(.id|test("-S[0-9]+$"))|. as $task|select($producers|index($task.id)|not)|"| \(.id) | なし（終端Task） | \(.deliverable) |"' <<<"$TASKS_JSON"
  printf '\n### 明示されたGate／Release接続\n\n| Milestone表現 | Consumer leaf | 依存時点 |\n| --- | --- | --- |\n'
  if [[ -n $milestones ]]; then jq -Rrsc 'split("\n")|map(select(length>0)|fromjson)|reduce .[] as $row ([]; if index($row) then . else .+[$row] end)|.[]|"| "+join(" | ")+" |"' <<<"$milestones"; else printf '| なし | — | — |\n'; fi
}

main() {
  local kind id number issue root_issue
  MODE=check REPO=$REPO_DEFAULT
  while (($#)); do
    case $1 in
      --check) MODE=check ;; --apply) MODE=apply ;; --verify) MODE=verify ;; --render-connections) MODE=render_connections ;;
      --repo) shift; (($#)) || abort_with '--repoに値が必要です'; REPO=$1 ;;
      -h|--help) printf 'Usage: task_issue_sync.sh [--check|--apply|--verify|--render-connections] [--repo OWNER/REPO]\n'; return ;;
      *) abort_with "未知の引数: $1" ;;
    esac
    shift
  done

  require_command jq 'Task MapとGitHub Issueの文字列をJSONとして安全に処理するため'
  require_supported_versions
  if [[ $MODE == apply || $MODE == verify ]]; then
    require_command gh 'GitHub Issueを作成・更新・照合するため'
  fi
  if [[ $MODE == render_connections ]]; then TASKS_JSON=$(read_tasks_from_file); validate_paths; render_connections; return; fi
  TASKS_JSON=$(read_tasks_from_file); validate_paths
  if [[ $MODE == check ]]; then printf 'CHECK OK L=6 M=19 S=116 TOTAL=141\n'; return; fi

  ISSUES_JSON=$(gh_json); MANAGED_JSON=$(managed_issues "$ISSUES_JSON"); title_conflicts
  MAPPING_JSON=$(jq 'map_values(.number)' <<<"$MANAGED_JSON")
  if [[ $MODE == verify ]]; then verify_all "$MANAGED_JSON"; return; fi
  [[ $MODE == apply ]] || abort_with "未知のmode: $MODE"

  for kind in L M S; do
    while IFS= read -r id; do
      [[ -n $(mapping_get "$id") ]] && continue
      create_issue "$id" "$(body_for "$id")"
      number=$CREATED_NUMBER
      MAPPING_JSON=$(jq --arg id "$id" --argjson number "$number" '.[$id]=$number' <<<"$MAPPING_JSON")
    done < <(jq -r --arg kind "$kind" '.[]|select((if (.id|test("-S[0-9]+$")) then "S" elif (.id|test("-M[0-9]+$")) then "M" else "L" end)==$kind)|.id' <<<"$TASKS_JSON")
  done

  ISSUES_JSON=$(gh_json); MANAGED_JSON=$(managed_issues "$ISSUES_JSON"); MAPPING_JSON=$(jq 'map_values(.number)' <<<"$MANAGED_JSON"); ISSUE_STATES_JSON=$(jq 'map_values(.state)' <<<"$MANAGED_JSON")
  while IFS= read -r id; do issue=$(jq --arg id "$id" '.[$id]' <<<"$MANAGED_JSON"); update_issue "$id" "$issue" "$(body_for "$id")"; done < <(jq -r '.[].id' <<<"$TASKS_JSON")
  ISSUES_JSON=$(gh_json); root_issue=$(jq --argjson number "$ROOT_ISSUE" '.[]|select(.number==$number)' <<<"$ISSUES_JSON"); [[ -n $root_issue ]] || abort_with "原典Issue #$ROOT_ISSUEがありません"; update_root_issue "$root_issue"
  ISSUES_JSON=$(gh_json); MANAGED_JSON=$(managed_issues "$ISSUES_JSON"); verify_all "$MANAGED_JSON"
}

if [[ ${BASH_SOURCE[0]} == "$0" ]]; then
  main "$@"
fi
