# ADR 0001: Go API With Chi And SQLite

Date: 2026-05-03

## Status

Accepted.

## Context

The product needs a small self-hosted backend that can serve a searchable calendar, export ICS, render print-friendly pages, and run comfortably under Docker Compose on a server. The data volume is small, but the app needs durable snapshots and auditable parser output.

## Decision

Use Go for a single deployable binary, `chi` for routing, and SQLite for the MVP database. The SQLite driver is the pure-Go `modernc.org/sqlite` package so local development and amd64 Docker builds do not depend on CGO.

## Consequences

The service stays simple to run and backup. SQLite is enough for Arad-scale lookup data and generated events. If the project expands to multiple counties or write-heavy workflows, the store boundary can grow a Postgres implementation without changing HTTP contracts.
