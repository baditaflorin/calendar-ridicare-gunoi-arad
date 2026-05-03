# ADR 0002: Offline ETL With Raw Snapshots

Date: 2026-05-03

## Status

Accepted.

## Context

RETIM pages are public HTML pages with large tables. Scraping live on every user search would make the user experience slow and fragile, and would make outages or HTML changes visible directly to citizens.

## Decision

Fetch official sources on a schedule, persist raw HTML snapshots with content hashes, parse into normalized SQLite tables, and serve searches from the local database. The app can bootstrap ETL on startup and refresh periodically, but user requests never scrape RETIM synchronously.

## Consequences

Users get fast responses, and maintainers get reproducible parser debugging. Runtime storage must include both the SQLite database and raw snapshot directory. Parser regressions can be fixed and replayed from stored source material.
