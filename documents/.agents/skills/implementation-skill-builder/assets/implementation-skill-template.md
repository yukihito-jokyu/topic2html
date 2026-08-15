---
name: impl-<scope>
description: <Front-load exactly what implementation work this skill owns, when it should trigger, and the closest adjacent work it should not own.>
---

# <Implementation Skill Name>

## Goal

<What repeatable implementation capability this skill provides.>

## Use When

- <trigger>

## Do Not Use When

- <adjacent scope owned elsewhere>
- <feature-specific one-off work that does not justify a skill>

## Required Inputs

- `features/{feature_id}/implementation-handoff.yaml`
- <relevant approved design>
- `AGENTS.md`
- <relevant existing code/conventions>

## Project Evidence

Rules in this skill are grounded in:

1. <AGENTS.md or explicit project rule>
2. <approved design/decision>
3. <dominant repository pattern>

Do not turn a single incidental example into a repository-wide rule.

## Owned Implementation Scope

- <logical responsibility>

## File Placement Rules

<Technology-specific paths and placement conventions may be defined here because this is the Implementation domain.>

## Autonomous Decisions

- <local/reversible implementation choices>

## Escalate When

- implementation requires a new Business Rule
- implementation requires changing Initial Design or a cross-feature contract
- implementation requires adopting a new library/framework not already approved
- implementation would violate an approved constraint

## Procedure

1. Read the implementation handoff and identify tasks owned by this skill.
2. Inspect the closest project conventions and existing analogous code.
3. Implement the smallest coherent change satisfying the task acceptance criteria.
4. Add or update tests at the appropriate boundary.
5. Run the validation commands below.
6. Report any design mismatch instead of silently changing planning artifacts.

## Validation

- `<build command>`
- `<test command>`
- `<lint command>`

## Completion Criteria

- Task acceptance criteria are satisfied.
- Project conventions are respected.
- Validation passes.
- No unapproved architecture/business decision was introduced.
