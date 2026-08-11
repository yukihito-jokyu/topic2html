#!/usr/bin/env bash
set -euo pipefail

PROGRAM_NAME="$(basename "$0")"
SKILL_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
HANDOFF_TEMPLATE="$SKILL_ROOT/references/session-handoff-template.md"
REQUIRED_SKILLS_PROMPT_HEADING='## Codex開始プロンプト（必須3スキル・版1）'
ISSUE_BODY_FILE=""

cleanup() {
  if [ -n "$ISSUE_BODY_FILE" ] && [ -f "$ISSUE_BODY_FILE" ]; then
    rm -f "$ISSUE_BODY_FILE"
  fi
}
trap cleanup EXIT

die() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

note() {
  printf '%s\n' "$*"
}

need_command() {
  command -v "$1" >/dev/null 2>&1 || die "必要なcommandがありません: $1"
}

usage() {
  cat <<USAGE
Usage:
  $PROGRAM_NAME plan ISSUE [--base REF] [--gate-commit SHA] [--fetch]
  $PROGRAM_NAME start ISSUE [--base REF] [--gate-commit SHA] [--no-fetch] [--no-open]
  $PROGRAM_NAME status [ISSUE]
  $PROGRAM_NAME open ISSUE
  $PROGRAM_NAME finish ISSUE
  $PROGRAM_NAME remove ISSUE --merged-into REF --confirm [--delete-branch]
USAGE
}

absolute_git_path() {
  raw_path="$1"
  base_path="$2"
  case "$raw_path" in
    /*) candidate="$raw_path" ;;
    *) candidate="$base_path/$raw_path" ;;
  esac
  candidate_dir="$(cd "$(dirname "$candidate")" && pwd -P)"
  printf '%s/%s\n' "$candidate_dir" "$(basename "$candidate")"
}

init_repository() {
  need_command git
  current_root="$(git rev-parse --show-toplevel 2>/dev/null)" || die "Git repository内で実行してください"
  current_root="$(cd "$current_root" && pwd -P)"
  raw_common_dir="$(git -C "$current_root" rev-parse --git-common-dir)"
  common_dir="$(absolute_git_path "$raw_common_dir" "$current_root")"
  primary_root="$(dirname "$common_dir")"
  primary_root="$(cd "$primary_root" && pwd -P)"
  worktree_root="$primary_root/.worktrees"
}

require_primary_checkout() {
  [ "$current_root" = "$primary_root" ] ||
    die "この操作はPrimary checkoutから実行してください: $primary_root"
}

load_repository_slug() {
  need_command gh
  repo_slug="$(gh repo view --json nameWithOwner --jq .nameWithOwner 2>/dev/null)" ||
    die "GitHub repositoryを特定できません。gh auth statusとoriginを確認してください"
}

validate_issue_number() {
  case "$1" in
    ''|*[!0-9]*) die "Issue番号は正の整数で指定してください: $1" ;;
    0) die "Issue番号は1以上で指定してください" ;;
  esac
}

load_issue() {
  issue_number="$1"
  validate_issue_number "$issue_number"
  load_repository_slug

  if [ -n "$ISSUE_BODY_FILE" ] && [ -f "$ISSUE_BODY_FILE" ]; then
    rm -f "$ISSUE_BODY_FILE"
  fi
  ISSUE_BODY_FILE="$(mktemp "${TMPDIR:-/tmp}/manage-task-worktrees.XXXXXX")"

  gh issue view "$issue_number" --repo "$repo_slug" --json body --jq .body >"$ISSUE_BODY_FILE" ||
    die "Issue #${issue_number}を取得できません"
  issue_title="$(gh issue view "$issue_number" --repo "$repo_slug" --json title --jq .title)"
  issue_state="$(gh issue view "$issue_number" --repo "$repo_slug" --json state --jq .state)"
  issue_url="$(gh issue view "$issue_number" --repo "$repo_slug" --json url --jq .url)"

  task_id="$(
    grep -Eo 'knowledge-task-id: L[0-9]+(-M[0-9]+)?(-S[0-9]+)?' "$ISSUE_BODY_FILE" |
      head -1 |
      awk '{print $2}'
  )"
  [ -n "$task_id" ] || die "Issue #${issue_number}にTask ID Markerがありません。初回起動準備としてIssue作成工程との接続を確認し、<!-- knowledge-task-id: Lx-My-Sz -->を同期してください"
  printf '%s\n' "$task_id" | grep -Eq '^L[0-9]+-M[0-9]+-S[0-9]+$' ||
    die "Issue #${issue_number}はtracking Taskです: ${task_id}。親Issueからleafへの展開手順に従ってください"

  task_slug="$(printf '%s' "$task_id" | tr '[:upper:]' '[:lower:]')"
  branch_name="task/issue-${issue_number}-${task_slug}"
  worktree_path="$worktree_root/issue-${issue_number}-${task_slug}"

  planning_snapshot="$(
    sed -nE 's/.*planning-snapshot: ([0-9a-f]{40}).*/\1/p' "$ISSUE_BODY_FILE" |
      head -1
  )"
  [ -n "$planning_snapshot" ] || die "Issue #${issue_number}にPlanning snapshot SHAがありません。初回起動準備として既存Commitを再利用できるか確認し、<!-- planning-snapshot: <40桁SHA> -->を同期してください"

  owner_paths="$(
    awk -F '|' '
      {
        key=$2
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", key)
        if (key=="書込み可能なPath／Glob") {
          value=$3
          gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
          print value
          exit
        }
      }
    ' "$ISSUE_BODY_FILE"
  )"
  [ -n "$owner_paths" ] || die "Issue #${issue_number}に書込み可能なPath／Globがありません"
  if printf '%s\n' "$owner_paths" | grep -Eq 'TBD|未確定'; then
    die "書込み可能なPath／Globに未解決TBDがあります: $owner_paths"
  fi
}

dependency_rows() {
  awk -F '|' '
    {
      key=$2
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", key)
      if (key=="着手依存" || key=="Gate通過依存") print $0
    }
  ' "$ISSUE_BODY_FILE"
}

dependency_issue_numbers() {
  dependency_rows |
    grep -Eo '(/issues/[0-9]+|#[0-9]+)' |
    sed -E 's#.*/##; s/^#//' |
    sort -nu || true
}

task_id_from_file() {
  grep -Eo 'knowledge-task-id: L[0-9]+(-M[0-9]+)?(-S[0-9]+)?' "$1" |
    head -1 |
    awk '{print $2}'
}

is_ancestor_task() {
  ancestor_task_id="$1"
  descendant_task_id="$2"
  case "$descendant_task_id" in
    "${ancestor_task_id}-"*) return 0 ;;
    *) return 1 ;;
  esac
}

gate_is_required() {
  gate_row="$(
    awk -F '|' '
      {
        key=$2
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", key)
        if (key=="Gate通過依存") print $0
      }
    ' "$ISSUE_BODY_FILE"
  )"
  [ -n "$gate_row" ] || return 1
  printf '%s\n' "$gate_row" | grep -Eq 'Gate' || return 1
  printf '%s\n' "$gate_row" | grep -Eq '該当なし' && return 1
  return 0
}

find_dependency_commit() {
  dep_body_file="$1"
  sed -n '/human-progress:start/,/human-progress:end/p' "$dep_body_file" |
    grep -Eo '[0-9a-fA-F]{7,40}' |
    tail -1 || true
}

validate_dependencies() {
  base_sha="$1"
  dependency_summary=""
  tracking_context_summary=""

  for dep_issue in $(dependency_issue_numbers); do
    dep_body_file="$(mktemp "${TMPDIR:-/tmp}/manage-task-dependency.XXXXXX")"
    gh issue view "$dep_issue" --repo "$repo_slug" --json body --jq .body >"$dep_body_file"
    dep_task_id="$(task_id_from_file "$dep_body_file")"
    [ -n "$dep_task_id" ] || die "依存Issue #${dep_issue}にTask ID Markerがありません"

    if is_ancestor_task "$dep_task_id" "$task_id"; then
      tracking_context_summary="${tracking_context_summary}#${dep_issue}:${dep_task_id} "
      rm -f "$dep_body_file"
      continue
    fi

    dep_state="$(gh issue view "$dep_issue" --repo "$repo_slug" --json state --jq .state)"
    [ "$dep_state" = "CLOSED" ] || die "着手依存Issue #${dep_issue}が未完了です: $dep_state"

    dep_commit="$(find_dependency_commit "$dep_body_file")"
    rm -f "$dep_body_file"

    [ -n "$dep_commit" ] || die "依存Issue #${dep_issue}のhuman-progressに統合Commitがありません"
    git -C "$primary_root" cat-file -e "${dep_commit}^{commit}" 2>/dev/null ||
      die "依存Issue #${dep_issue}のCommitをlocalで解決できません: $dep_commit"
    git -C "$primary_root" merge-base --is-ancestor "$dep_commit" "$base_sha" ||
      die "依存Issue #${dep_issue}のCommit $dep_commit が基準refに含まれていません"
    dependency_summary="${dependency_summary}#${dep_issue}:${dep_commit} "
  done

  [ -n "$dependency_summary" ] || dependency_summary="該当なし"
  [ -n "$tracking_context_summary" ] || tracking_context_summary="該当なし"
}

owner_path_lines() {
  raw_owner_paths="$1"
  raw_owner_paths="${raw_owner_paths//\`/}"
  raw_owner_paths="${raw_owner_paths//、/$'\n'}"
  raw_owner_paths="${raw_owner_paths//,/$'\n'}"
  printf '%s\n' "$raw_owner_paths" |
    sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//' |
    sed '/^$/d'
}

simple_owner_path() {
  printf '%s\n' "$1" | grep -Eq '^[[:alnum:]_.\/-]+(/\*\*)?$'
}

simple_paths_overlap() {
  first_path="$1"
  second_path="$2"
  simple_owner_path "$first_path" && simple_owner_path "$second_path" || return 2

  first_base="${first_path%/\*\*}"
  second_base="${second_path%/\*\*}"
  [ "$first_base" = "$second_base" ] && return 0
  case "$first_base" in "$second_base"/*) return 0 ;; esac
  case "$second_base" in "$first_base"/*) return 0 ;; esac
  return 1
}

validate_active_path_conflicts() {
  path_audit_summary="機械判定合格。複雑Globと共有資産は手動確認が必要"

  while IFS= read -r active_worktree; do
    [ "$active_worktree" = "$primary_root" ] && continue
    [ "$active_worktree" = "$worktree_path" ] && continue
    active_handoff="$active_worktree/.codex/task-session.local.md"
    [ -f "$active_handoff" ] || continue
    active_owner_paths="$(sed -n 's/^- 書込み可能なPath／Glob: //p' "$active_handoff" | head -1)"
    [ -n "$active_owner_paths" ] || continue

    while IFS= read -r candidate_path; do
      while IFS= read -r active_path; do
        if simple_paths_overlap "$candidate_path" "$active_path"; then
          die "既存worktreeとOwner Pathが競合します: $candidate_path <-> $active_path ($active_worktree)"
        fi
      done < <(owner_path_lines "$active_owner_paths")
    done < <(owner_path_lines "$owner_paths")
  done < <(git -C "$primary_root" worktree list --porcelain | sed -n 's/^worktree //p')
}

validate_ignore_rules() {
  git -C "$primary_root" check-ignore -q --no-index .worktrees/.probe ||
    die "$primary_root/.gitignoreに /.worktrees/ がありません。初回起動準備として管理機能導入用branchでignore規則をbaseへ追加してください"
  base_ignore="$(git -C "$primary_root" show "${base_sha}:.gitignore" 2>/dev/null)" ||
    die "基準refに.gitignoreがありません: $base_ref"
  printf '%s\n' "$base_ignore" | grep -Fxq '/.worktrees/' ||
    die "基準refの.gitignoreに /.worktrees/ がありません。初回起動準備として管理機能導入用branchをbaseへ統合してください"
  printf '%s\n' "$base_ignore" | grep -Fxq '/.codex/task-session.local.md' ||
    die "基準refの.gitignoreに /.codex/task-session.local.md がありません。初回起動準備として管理機能導入用branchをbaseへ統合してください"
}

validate_ref_required_skills() {
  required_ref="$1"
  required_ref_label="$2"
  for required_skill in \
    .agents/skills/manage-task-worktrees/SKILL.md \
    .agents/skills/conduct-task-discussion/SKILL.md \
    .agents/skills/explain-with-context/SKILL.md; do
    git -C "$primary_root" cat-file -e "${required_ref}:${required_skill}" 2>/dev/null ||
      die "${required_ref_label}にTask実行で必須のSkillがありません: $required_skill"
  done
}

validate_worktree_required_skills() {
  for required_skill in \
    .agents/skills/manage-task-worktrees/SKILL.md \
    .agents/skills/conduct-task-discussion/SKILL.md \
    .agents/skills/explain-with-context/SKILL.md; do
    [ -f "$worktree_path/$required_skill" ] ||
      die "worktreeにTask実行で必須のSkillがありません。基準refを安全に統合してから再開してください: $required_skill"
  done
}

registered_path_for_branch() {
  git -C "$primary_root" worktree list --porcelain |
    awk -v wanted="refs/heads/$branch_name" '
      /^worktree / {path=substr($0, 10)}
      /^branch / {
        ref=substr($0, 8)
        if (ref==wanted) print path
      }
    '
}

print_plan() {
  note "Task:       $task_id / Issue #$issue_number"
  note "Title:      $issue_title"
  note "Repository: $repo_slug"
  note "Base ref:   $base_ref"
  note "Base SHA:   $base_sha"
  note "Gate SHA:   ${gate_commit:-該当なし}"
  note "Depends:    $dependency_summary"
  note "Context:    $tracking_context_summary"
  note "Branch:     $branch_name"
  note "Worktree:   $worktree_path"
  note "Owner:      $owner_paths"
  note "Path audit: $path_audit_summary"
  note "Snapshot:   $planning_snapshot"
}

escape_awk_replacement() {
  printf '%s' "$1" | sed 's/[\\&]/\\&/g'
}

render_handoff() {
  [ -f "$HANDOFF_TEMPLATE" ] || die "引継ぎtemplateがありません: $HANDOFF_TEMPLATE"
  handoff_dir="$worktree_path/.codex"
  handoff_path="$handoff_dir/task-session.local.md"
  mkdir -p "$handoff_dir"

  git -C "$worktree_path" check-ignore -q --no-index .codex/task-session.local.md ||
    die "handoff fileがignoreされません: $handoff_path"

  if [ -e "$handoff_path" ]; then
    [ -f "$handoff_path" ] || die "handoff fileの配置先が通常ファイルではありません: $handoff_path"
    if ! grep -Fxq "$REQUIRED_SKILLS_PROMPT_HEADING" "$handoff_path"; then
      cat >>"$handoff_path" <<'HANDOFF_APPEND'

## Codex開始プロンプト（必須3スキル・版1）

既存の質問、回答、判断履歴を保持したまま、次の指示を追加する。

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
HANDOFF_APPEND
      note "UPDATED: 既存handoffの履歴を保持し、必須Skillの開始プロンプトを追記しました"
    else
      note "PRESERVED: 既存handoffと質問・回答・判断履歴を変更しません"
    fi
    return 0
  fi

  safe_title="$(escape_awk_replacement "$issue_title")"
  safe_owner="$(escape_awk_replacement "$owner_paths")"

  awk \
    -v issue_number="$issue_number" \
    -v issue_url="$issue_url" \
    -v task_id="$task_id" \
    -v issue_title="$safe_title" \
    -v branch="$branch_name" \
    -v worktree="$worktree_path" \
    -v base_ref="$base_ref" \
    -v base_sha="$base_sha" \
    -v gate_commit="${gate_commit:-該当なし}" \
    -v snapshot="$planning_snapshot" \
    -v owner="$safe_owner" '
      {
        gsub(/{{ISSUE_NUMBER}}/, issue_number)
        gsub(/{{ISSUE_URL}}/, issue_url)
        gsub(/{{TASK_ID}}/, task_id)
        gsub(/{{ISSUE_TITLE}}/, issue_title)
        gsub(/{{BRANCH}}/, branch)
        gsub(/{{WORKTREE_PATH}}/, worktree)
        gsub(/{{BASE_REF}}/, base_ref)
        gsub(/{{BASE_SHA}}/, base_sha)
        gsub(/{{GATE_COMMIT}}/, gate_commit)
        gsub(/{{PLANNING_SNAPSHOT}}/, snapshot)
        gsub(/{{OWNER_PATHS}}/, owner)
        print
      }
    ' "$HANDOFF_TEMPLATE" >"$handoff_path"
}

parse_start_options() {
  issue_arg=""
  base_ref="origin/main"
  gate_commit=""
  fetch_requested="$1"
  open_requested=true
  shift

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --base)
        [ "$#" -ge 2 ] || die "--baseには値が必要です"
        base_ref="$2"
        shift 2
        ;;
      --gate-commit)
        [ "$#" -ge 2 ] || die "--gate-commitには値が必要です"
        gate_commit="$2"
        shift 2
        ;;
      --fetch)
        fetch_requested=true
        shift
        ;;
      --no-fetch)
        fetch_requested=false
        shift
        ;;
      --no-open)
        open_requested=false
        shift
        ;;
      --*)
        die "不明なoptionです: $1"
        ;;
      *)
        [ -z "$issue_arg" ] || die "Issue番号は1つだけ指定してください"
        issue_arg="$1"
        shift
        ;;
    esac
  done
  [ -n "$issue_arg" ] || die "Issue番号が必要です"
}

run_plan_or_start() {
  action="$1"
  shift
  [ "$action" = "start" ] && default_fetch=true || default_fetch=false
  parse_start_options "$default_fetch" "$@"

  init_repository
  require_primary_checkout
  load_issue "$issue_arg"
  [ "$issue_state" = "OPEN" ] || die "Issue #${issue_number}はOPENではありません: $issue_state"

  if [ "$fetch_requested" = true ]; then
    git -C "$primary_root" fetch --prune origin
  fi

  base_sha="$(git -C "$primary_root" rev-parse --verify "${base_ref}^{commit}" 2>/dev/null)" ||
    die "基準refを解決できません: $base_ref"

  git -C "$primary_root" cat-file -e "${planning_snapshot}^{commit}" 2>/dev/null ||
    die "Planning snapshotをlocalで解決できません: $planning_snapshot"
  git -C "$primary_root" merge-base --is-ancestor "$planning_snapshot" "$base_sha" ||
    die "Planning snapshot $planning_snapshot が基準ref $base_ref に含まれていません"

  validate_ref_required_skills "$base_sha" "基準ref"
  validate_ignore_rules
  validate_dependencies "$base_sha"
  validate_active_path_conflicts

  if gate_is_required; then
    [ -n "$gate_commit" ] || die "Gate通過依存があります。--gate-commit <sha>を指定してください"
  fi
  if [ -n "$gate_commit" ]; then
    gate_sha="$(git -C "$primary_root" rev-parse --verify "${gate_commit}^{commit}" 2>/dev/null)" ||
      die "Gate Commitを解決できません: $gate_commit"
    git -C "$primary_root" merge-base --is-ancestor "$gate_sha" "$base_sha" ||
      die "Gate Commit $gate_sha が基準ref $base_ref に含まれていません"
    gate_commit="$gate_sha"
  fi

  print_plan
  [ "$action" = "plan" ] && return 0
  [ "$open_requested" = false ] || need_command code

  existing_branch_path="$(registered_path_for_branch)"
  if [ -n "$existing_branch_path" ] && [ "$existing_branch_path" != "$worktree_path" ]; then
    die "${branch_name}は別worktreeで使用中です: $existing_branch_path"
  fi

  if [ -e "$worktree_path" ]; then
    git -C "$worktree_path" rev-parse --is-inside-work-tree >/dev/null 2>&1 ||
      die "配置先が既存の非worktree Pathです: $worktree_path"
    actual_branch="$(git -C "$worktree_path" branch --show-current)"
    [ "$actual_branch" = "$branch_name" ] || die "既存worktreeのbranchが一致しません: $actual_branch"
    validate_ref_required_skills "$branch_name" "既存branch"
    note "RESUME: $worktree_path"
  else
    if git -C "$primary_root" show-ref --verify --quiet "refs/heads/$branch_name"; then
      validate_ref_required_skills "$branch_name" "既存branch"
      mkdir -p "$worktree_root"
      git -C "$primary_root" worktree add "$worktree_path" "$branch_name"
    elif git -C "$primary_root" show-ref --verify --quiet "refs/remotes/origin/$branch_name"; then
      validate_ref_required_skills "origin/$branch_name" "既存remote branch"
      mkdir -p "$worktree_root"
      git -C "$primary_root" worktree add --track -b "$branch_name" "$worktree_path" "origin/$branch_name"
    else
      mkdir -p "$worktree_root"
      git -C "$primary_root" worktree add -b "$branch_name" "$worktree_path" "$base_sha"
    fi
    note "CREATED: $worktree_path"
  fi

  validate_worktree_required_skills
  render_handoff
  note "HANDOFF: $handoff_path"

  if [ "$open_requested" = true ]; then
    code --new-window "$worktree_path" "$handoff_path"
    note "VS Codeを開きました。Codexを手動開始し、handoffの開始promptを渡してください。"
  else
    note "VS Codeは開いていません。open commandで後から開けます。"
  fi
}

run_status() {
  init_repository
  issue_arg="${1:-}"
  if [ -z "$issue_arg" ]; then
    git -C "$primary_root" worktree list
    return 0
  fi
  [ "$#" -eq 1 ] || die "statusのIssue番号は1つだけ指定してください"
  load_issue "$issue_arg"
  if [ ! -e "$worktree_path" ]; then
    note "NOT CREATED: $worktree_path"
    return 0
  fi
  note "Task: $task_id / Issue #$issue_number"
  note "Path: $worktree_path"
  git -C "$worktree_path" status --short --branch
  note "HEAD: $(git -C "$worktree_path" rev-parse HEAD)"
}

run_open() {
  [ "$#" -eq 1 ] || die "openにはIssue番号が必要です"
  init_repository
  load_issue "$1"
  [ -d "$worktree_path" ] || die "worktreeがありません: $worktree_path"
  need_command code
  handoff_path="$worktree_path/.codex/task-session.local.md"
  [ -f "$handoff_path" ] || die "handoff fileがありません: $handoff_path"
  actual_branch="$(git -C "$worktree_path" branch --show-current)"
  [ "$actual_branch" = "$branch_name" ] || die "既存worktreeのbranchが一致しません: $actual_branch"
  validate_worktree_required_skills
  render_handoff
  code --new-window "$worktree_path" "$handoff_path"
}

run_finish() {
  [ "$#" -eq 1 ] || die "finishにはIssue番号が必要です"
  init_repository
  load_issue "$1"
  [ -d "$worktree_path" ] || die "worktreeがありません: $worktree_path"

  actual_branch="$(git -C "$worktree_path" branch --show-current)"
  [ "$actual_branch" = "$branch_name" ] || die "worktreeのbranchが一致しません: $actual_branch"

  note "Task:   $task_id / Issue #$issue_number"
  note "Branch: $branch_name"
  note "HEAD:   $(git -C "$worktree_path" rev-parse HEAD)"
  note "Owner:  $owner_paths"
  note ""
  note "Git status:"
  git -C "$worktree_path" status --short --branch

  changed_paths="$(
    {
      git -C "$worktree_path" diff --name-only
      git -C "$worktree_path" diff --cached --name-only
      git -C "$worktree_path" ls-files --others --exclude-standard
    } | sort -u
  )"
  note ""
  note "Changed paths:"
  [ -z "$changed_paths" ] && note "該当なし" || printf '%s\n' "$changed_paths"

  if git -C "$worktree_path" rev-parse --abbrev-ref '@{upstream}' >/dev/null 2>&1; then
    upstream="$(git -C "$worktree_path" rev-parse --abbrev-ref '@{upstream}')"
    note ""
    note "Upstream: $upstream"
    git -C "$worktree_path" rev-list --left-right --count "$upstream...HEAD"
  else
    note ""
    note "Upstream: 未設定"
  fi

  note ""
  note "この出力だけでは完了になりません。Issue固有テスト、Owner境界、PR、統合Evidenceを確認してください。"
}

run_remove() {
  issue_arg=""
  merged_into=""
  confirmed=false
  delete_branch=false

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --merged-into)
        [ "$#" -ge 2 ] || die "--merged-intoには値が必要です"
        merged_into="$2"
        shift 2
        ;;
      --confirm)
        confirmed=true
        shift
        ;;
      --delete-branch)
        delete_branch=true
        shift
        ;;
      --*) die "不明なoptionです: $1" ;;
      *)
        [ -z "$issue_arg" ] || die "Issue番号は1つだけ指定してください"
        issue_arg="$1"
        shift
        ;;
    esac
  done

  [ -n "$issue_arg" ] || die "Issue番号が必要です"
  [ -n "$merged_into" ] || die "--merged-into <ref>が必要です"
  [ "$confirmed" = true ] || die "削除には--confirmが必要です"

  init_repository
  require_primary_checkout
  load_issue "$issue_arg"
  [ -d "$worktree_path" ] || die "worktreeがありません: $worktree_path"

  actual_branch="$(git -C "$worktree_path" branch --show-current)"
  [ "$actual_branch" = "$branch_name" ] || die "worktreeのbranchが一致しません: $actual_branch"
  [ -z "$(git -C "$worktree_path" status --porcelain)" ] || die "worktreeに未Commit変更があります"

  head_sha="$(git -C "$worktree_path" rev-parse HEAD)"
  merged_sha="$(git -C "$primary_root" rev-parse --verify "${merged_into}^{commit}" 2>/dev/null)" ||
    die "統合先refを解決できません: $merged_into"
  git -C "$primary_root" merge-base --is-ancestor "$head_sha" "$merged_sha" ||
    die "worktree HEAD $head_sha は $merged_into にMergeされていません"

  git -C "$primary_root" worktree remove "$worktree_path"
  note "REMOVED: $worktree_path"

  if [ "$delete_branch" = true ]; then
    git -C "$primary_root" branch -d "$branch_name"
    note "DELETED BRANCH: $branch_name"
  else
    note "Branchは保持しました: $branch_name"
  fi
}

command_name="${1:-help}"
[ "$#" -eq 0 ] || shift

case "$command_name" in
  plan) run_plan_or_start plan "$@" ;;
  start) run_plan_or_start start "$@" ;;
  status) run_status "$@" ;;
  open) run_open "$@" ;;
  finish) run_finish "$@" ;;
  remove) run_remove "$@" ;;
  help|-h|--help) usage ;;
  *)
    usage >&2
    die "不明なcommandです: $command_name"
    ;;
esac
