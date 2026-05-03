# ADR 0003: Local Hooks Instead Of GitHub Actions

Date: 2026-05-03

## Status

Accepted.

## Context

The project owner explicitly asked for checks that can be run locally and not through GitHub Actions.

## Decision

Keep executable hooks under `.githooks/` and use `scripts/install-hooks.sh` to configure `core.hooksPath`. The hooks run formatting, vetting, unit tests, and smoke tests before commits and pushes.

## Consequences

Checks run before code leaves the workstation. Contributors must install hooks once after cloning. CI can be added later if desired, but this repository intentionally does not include GitHub Actions.
