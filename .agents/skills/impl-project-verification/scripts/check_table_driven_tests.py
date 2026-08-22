#!/usr/bin/env python3
import argparse
from dataclasses import dataclass
from pathlib import Path
import re
import sys


@dataclass(frozen=True)
class Candidate:
    path: Path
    name: str
    subtests: int
    checks: int


FUNCTION = re.compile(r"^func (Test\w+)\s*\(", re.MULTILINE)
SUBTEST = re.compile(r"\bt\.Run\s*\(")
CHECK = re.compile(r"\bif\s+")
RANGE = re.compile(r"\bfor\b[^\n{]*\brange\b")
CASE_NAME = re.compile(r'^\s*name:\s*"', re.MULTILINE)


def table_driven(body: str) -> bool:
    if RANGE.search(body) is None:
        return False
    if len(SUBTEST.findall(body)) >= 2:
        return True
    if "[]struct" not in body and "range []" not in body:
        return False
    return len(CASE_NAME.findall(body)) >= 2


def inspect(root: Path) -> list[Candidate]:
    candidates: list[Candidate] = []
    for path in sorted(root.rglob("*_test.go")):
        source = path.read_text(encoding="utf-8")
        functions = list(FUNCTION.finditer(source))
        for index, match in enumerate(functions):
            end = functions[index + 1].start() if index + 1 < len(functions) else len(source)
            body = source[match.end():end]
            subtests = len(SUBTEST.findall(body))
            checks = len(CHECK.findall(body))
            if not table_driven(body) and (subtests > 1 or checks > 3):
                candidates.append(Candidate(path, match.group(1), subtests, checks))
    return candidates


def main() -> int:
    parser = argparse.ArgumentParser(description="Goテストのテーブル駆動化の未対応候補を表示する")
    parser.add_argument("--root", default="backend", type=Path, help="Goテストを探索するディレクトリ")
    arguments = parser.parse_args()
    if not arguments.root.is_dir():
        parser.error(f"ディレクトリがありません: {arguments.root}")
    candidates = inspect(arguments.root)
    if not candidates:
        print("テーブル駆動化の未対応候補はありません。")
        return 0
    print("テーブル駆動化の未対応候補:")
    for candidate in candidates:
        print(f"- {candidate.path}: {candidate.name}（subtest: {candidate.subtests}、条件分岐: {candidate.checks}）")
    print("注記: 静的な候補検出です。単一シナリオの並行・資源管理テストは個別実装が適切な場合があります。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
