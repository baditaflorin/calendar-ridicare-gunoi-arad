# ADR 0001: Go API With Chi And SQLite

Date: 2026-05-03

## Status

Accepted.

## Context

The product needs a small self-hosted backend that can serve a searchable calendar, export ICS, render print-friendly pages, and run comfortably under Docker Compose on a server. The data volume is small, but the app needs durable snapshots and auditable parser output.

## Decision

Use Go for a single deployable binary, `chi` for routing, and SQLite for the MVP database. The SQLite driver is `github.com/mattn/go-sqlite3`, the most battle-tested Go SQLite driver. Docker builds install a compiler toolchain inside the builder stage so CGO stays contained in the image build.

## Consequences

The service stays simple to run and backup. SQLite is enough for Arad-scale lookup data and generated events. Local and Docker builds need CGO support. If the project expands to multiple counties or write-heavy workflows, the store boundary can grow a Postgres implementation without changing HTTP contracts.
