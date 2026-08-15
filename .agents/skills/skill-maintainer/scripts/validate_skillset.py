#!/usr/bin/env python3
"""Lightweight structural validation for this skillset.

Checks only deterministic structure. It does not evaluate model behavior.
"""
from __future__ import annotations

import sys
from pathlib import Path

try:
    import yaml
except ImportError:
    print("ERROR: PyYAML is required to run this validator.", file=sys.stderr)
    raise SystemExit(2)


def load_frontmatter(path: Path) -> dict:
    text = path.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        raise ValueError("SKILL.md must start with YAML frontmatter")
    parts = text.split("---", 2)
    if len(parts) != 3:
        raise ValueError("invalid YAML frontmatter delimiters")
    data = yaml.safe_load(parts[1]) or {}
    return data


def resolve_project_paths(repo: Path) -> tuple[list[Path], Path | None]:
    """Resolve the split implementation and documentation roots of this project."""
    direct_workflow = repo / ".ai" / "workflow"
    documents_dir = repo / "documents"
    documents_workflow = documents_dir / ".ai" / "workflow"

    if direct_workflow.exists():
        skill_roots = [repo / ".agents" / "skills"]
        parent_skill_root = repo.parent / ".agents" / "skills"
        if repo.name == "documents" and parent_skill_root.exists():
            skill_roots.append(parent_skill_root)
        return skill_roots, direct_workflow

    if documents_workflow.exists():
        return [repo / ".agents" / "skills", documents_dir / ".agents" / "skills"], documents_workflow

    return [repo / ".agents" / "skills"], None


def direct_child_skill_files(skill_roots: list[Path]) -> list[Path]:
    """Collect direct-child skills once even if another root links to the same skill."""
    files: list[Path] = []
    seen: set[Path] = set()
    for root in skill_roots:
        for path in sorted(root.glob("*/SKILL.md")):
            resolved = path.resolve()
            if resolved not in seen:
                seen.add(resolved)
                files.append(path)
    return files


def validate_skill_scenarios(repo: Path, names: dict[str, Path], errors: list[str]) -> None:
    """Validate regression scenario ownership and the minimum deterministic schema."""
    scenario_root = repo / ".ai" / "skill-tests"
    required_keys = {
        "id",
        "target_skill",
        "scenario",
        "input_summary",
        "expected",
        "forbidden",
        "related_issue",
    }
    scenario_ids: dict[str, Path] = {}
    targets_with_scenarios: set[str] = set()

    if not scenario_root.exists():
        errors.append("Missing .ai/skill-tests regression scenario root")
        return

    for path in sorted(scenario_root.glob("*/*.yaml")):
        try:
            data = yaml.safe_load(path.read_text(encoding="utf-8")) or {}
        except Exception as exc:  # noqa: BLE001
            errors.append(f"{path}: invalid YAML: {exc}")
            continue
        if not isinstance(data, dict):
            errors.append(f"{path}: scenario must be a YAML mapping")
            continue

        missing = sorted(required_keys - data.keys())
        if missing:
            errors.append(f"{path}: missing scenario keys {', '.join(missing)}")

        scenario_id = data.get("id")
        if not isinstance(scenario_id, str) or not scenario_id.strip():
            errors.append(f"{path}: id must be a non-empty string")
        elif scenario_id in scenario_ids:
            errors.append(f"duplicate scenario id '{scenario_id}': {scenario_ids[scenario_id]} and {path}")
        else:
            scenario_ids[scenario_id] = path

        target = data.get("target_skill")
        if not isinstance(target, str) or not target.strip():
            errors.append(f"{path}: target_skill must be a non-empty string")
        else:
            targets_with_scenarios.add(target)
            if target not in names:
                errors.append(f"{path}: target_skill '{target}' does not exist")
            if path.parent.name != target:
                errors.append(
                    f"{path}: scenario directory '{path.parent.name}' != target_skill '{target}'"
                )

        for key in ("scenario", "input_summary", "related_issue"):
            if not isinstance(data.get(key), str) or not data[key].strip():
                errors.append(f"{path}: {key} must be a non-empty string")
        for key in ("expected", "forbidden"):
            values = data.get(key)
            if not isinstance(values, list) or not values or not all(
                isinstance(value, str) and value.strip() for value in values
            ):
                errors.append(f"{path}: {key} must be a non-empty list of strings")

    orchestrated_roles = {
        "impl-knowledge-cli",
        "impl-knowledge-cli-implementation",
        "verify-knowledge-cli-spec",
        "review-knowledge-cli",
        "audit-knowledge-cli-conformance",
    }
    for role in sorted(orchestrated_roles):
        if role in names and role not in targets_with_scenarios:
            errors.append(f"role skill '{role}' requires at least one regression scenario")


def main() -> int:
    repo = Path.cwd()
    skill_roots, workflow_root = resolve_project_paths(repo)
    errors: list[str] = []
    warnings: list[str] = []

    skill_files = direct_child_skill_files(skill_roots)
    if not skill_files:
        errors.append("No direct-child skills found under the configured .agents/skills roots")

    names: dict[str, Path] = {}
    for path in skill_files:
        try:
            fm = load_frontmatter(path)
        except Exception as exc:  # noqa: BLE001
            errors.append(f"{path}: {exc}")
            continue

        name = fm.get("name")
        description = fm.get("description")
        if not isinstance(name, str) or not name.strip():
            errors.append(f"{path}: missing non-empty frontmatter name")
        elif name != path.parent.name:
            errors.append(f"{path}: frontmatter name '{name}' != directory '{path.parent.name}'")
        elif name in names:
            errors.append(f"duplicate skill name '{name}': {names[name]} and {path}")
        else:
            names[name] = path

        if not isinstance(description, str) or not description.strip():
            errors.append(f"{path}: missing non-empty frontmatter description")
        elif len(description) > 1024:
            errors.append(f"{path}: description exceeds 1024 characters")

    direct_resolved = {p.resolve() for p in skill_files}
    nested: list[Path] = []
    seen_nested: set[Path] = set()
    for root in skill_roots:
        for path in root.rglob("SKILL.md"):
            resolved = path.resolve()
            if resolved not in direct_resolved and resolved not in seen_nested:
                seen_nested.add(resolved)
                nested.append(path)
    for path in nested:
        warnings.append(
            f"Nested SKILL.md detected: {path}. Confirm it is intended to be independently discovered."
        )

    orchestration_skill = names.get("impl-knowledge-cli")
    if orchestration_skill:
        orchestration_text = orchestration_skill.read_text(encoding="utf-8")
        required_roles = (
            "impl-knowledge-cli-implementation",
            "review-knowledge-cli",
            "verify-knowledge-cli-spec",
            "audit-knowledge-cli-conformance",
        )
        for role in required_roles:
            if role not in names:
                errors.append(f"impl-knowledge-cli requires missing role skill '{role}'")
            elif role not in orchestration_text:
                errors.append(f"impl-knowledge-cli does not delegate to role skill '{role}'")
        for role in (
            "review-knowledge-cli",
            "verify-knowledge-cli-spec",
            "audit-knowledge-cli-conformance",
        ):
            role_text = names[role].read_text(encoding="utf-8") if role in names else ""
            if "読み取り専用" not in role_text:
                errors.append(f"role skill '{role}' must declare read-only operation")

        contract = orchestration_skill.parent / "references" / "orchestration-contract.md"
        fingerprint = orchestration_skill.parent / "scripts" / "candidate_fingerprint.py"
        source_fingerprint = orchestration_skill.parent / "scripts" / "source_fingerprint.py"
        implementation_skill = names.get("impl-knowledge-cli-implementation")
        implementation_scripts = implementation_skill.parent / "scripts" if implementation_skill else None
        coverage_check = implementation_scripts / "check_test_coverage.py" if implementation_scripts else None
        literal_check = (
            implementation_scripts / "check_composite_literal_layout.py"
            if implementation_scripts
            else None
        )
        if not contract.is_file():
            errors.append("impl-knowledge-cli requires references/orchestration-contract.md")
        if not fingerprint.is_file():
            errors.append("impl-knowledge-cli requires scripts/candidate_fingerprint.py")
        if not source_fingerprint.is_file():
            errors.append("impl-knowledge-cli requires scripts/source_fingerprint.py")
        if not coverage_check or not coverage_check.is_file():
            errors.append(
                "impl-knowledge-cli-implementation requires scripts/check_test_coverage.py"
            )
        if not literal_check or not literal_check.is_file():
            errors.append(
                "impl-knowledge-cli-implementation requires scripts/check_composite_literal_layout.py"
            )

        for role_name in ("impl-knowledge-cli", *required_roles):
            if role_name not in names:
                continue
            scripts_dir = names[role_name].parent / "scripts"
            if not scripts_dir.exists():
                continue
            for script in sorted(path for path in scripts_dir.rglob("*") if path.is_file()):
                if script.suffix != ".py":
                    errors.append(
                        f"{role_name}: Skill helper scripts must use .py: {script.name}"
                    )

        if implementation_skill:
            implementation_text = implementation_skill.read_text(encoding="utf-8")
            if "File and Script Placement Rules" not in implementation_text:
                errors.append(
                    "impl-knowledge-cli-implementation must define file and script placement rules"
                )
            if "hidden fileを含むrepository全体" not in implementation_text:
                errors.append(
                    "impl-knowledge-cli-implementation must require repository-wide caller checks"
                )
        reviewer = names.get("review-knowledge-cli")
        if reviewer:
            reviewer_text = reviewer.read_text(encoding="utf-8")
            reviewer_coverage = reviewer.parent / "scripts" / "review_test_coverage.py"
            if "file配置matrix" not in reviewer_text:
                errors.append("review-knowledge-cli must review the file placement matrix")
            if "hidden fileを含むrepository全体" not in reviewer_text:
                errors.append(
                    "review-knowledge-cli must perform repository-wide caller checks"
                )
            if not reviewer_coverage.is_file():
                errors.append(
                    "review-knowledge-cli requires scripts/review_test_coverage.py"
                )
            elif "review_test_coverage.py" not in reviewer_text:
                errors.append(
                    "review-knowledge-cli must invoke its independent coverage script"
                )

        ci_workflow = repo / ".github" / "workflows" / "golangci-lint.yml"
        if not ci_workflow.is_file():
            errors.append("Missing .github/workflows/golangci-lint.yml")
        else:
            ci_text = ci_workflow.read_text(encoding="utf-8")
            required_ci_scripts = (
                ".agents/skills/impl-knowledge-cli-implementation/scripts/"
                "check_composite_literal_layout.py",
                ".agents/skills/impl-knowledge-cli-implementation/scripts/"
                "check_test_coverage.py",
            )
            for required_script in required_ci_scripts:
                if f"python3 {required_script}" not in ci_text:
                    errors.append(
                        f"golangci-lint workflow must call 'python3 {required_script}'"
                    )
            for obsolete_script in (
                "check_composite_literal_layout.sh",
                "check_composite_literal_layout.go",
                "check_test_coverage.sh",
            ):
                if obsolete_script in ci_text:
                    errors.append(
                        f"golangci-lint workflow references obsolete script '{obsolete_script}'"
                    )

        contract_roles = ("impl-knowledge-cli", *required_roles)
        for role in contract_roles:
            if role not in names:
                continue
            role_text = names[role].read_text(encoding="utf-8")
            if "orchestration-contract.md" not in role_text:
                errors.append(f"role skill '{role}' must reference orchestration-contract.md")
            metadata = names[role].parent / "agents" / "openai.yaml"
            if not metadata.is_file():
                errors.append(f"role skill '{role}' requires agents/openai.yaml")

    validate_skill_scenarios(repo, names, errors)

    artifact_map_path = workflow_root / "artifact-map.yaml" if workflow_root else None
    if artifact_map_path and artifact_map_path.exists():
        try:
            artifact_map = yaml.safe_load(artifact_map_path.read_text(encoding="utf-8")) or {}
            for artifact_name, spec in (artifact_map.get("artifacts") or {}).items():
                owner = (spec or {}).get("owner")
                if owner and owner not in names and owner != "implementation-skill-builder":
                    errors.append(
                        f"artifact '{artifact_name}' references unknown owner skill '{owner}'"
                    )
        except Exception as exc:  # noqa: BLE001
            errors.append(f"{artifact_map_path}: invalid YAML: {exc}")
    else:
        errors.append("Missing workflow artifact-map.yaml at this repository entry point")

    if workflow_root:
        for path in workflow_root.glob("*.yaml"):
            try:
                yaml.safe_load(path.read_text(encoding="utf-8"))
            except Exception as exc:  # noqa: BLE001
                errors.append(f"{path}: invalid YAML: {exc}")

    for warning in warnings:
        print(f"WARN: {warning}")
    for error in errors:
        print(f"ERROR: {error}")

    if errors:
        print(f"FAILED: {len(errors)} error(s), {len(warnings)} warning(s)")
        return 1

    print(f"OK: {len(skill_files)} skills, {len(warnings)} warning(s)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
