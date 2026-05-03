# ADR 0007: Static GitHub Pages Frontend

Date: 2026-05-03

## Status

Accepted.

## Context

The public citizen-facing product should be cheap, fast, and resilient. A live backend is useful for ETL, observability, Docker deployment, and admin workflows, but it is not required for serving a read-only calendar once the official RETIM data has been normalized.

## Decision

Add a static export command that writes a GitHub Pages bundle into `docs/`. The export includes HTML/CSS/JS assets, `data/catalog.json` for neighborhoods and streets, and one JSON file per place with generated events and source links. The frontend switches into static mode with `window.GUNOI_STATIC = true` and reads relative files instead of `/api/*`.

## Consequences

The public app can run from GitHub Pages with no server. Refreshing data becomes a local workflow: run ETL, export static files, commit, and push. The Docker backend remains available for self-hosted deployments and for generating the static bundle. Large raw HTML snapshots and SQLite databases stay ignored; only normalized public JSON is committed.
