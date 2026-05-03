# ADR 0005: Prometheus Metrics

Date: 2026-05-03

## Status

Accepted.

## Context

The scraper/parser is the riskiest part of the system because upstream HTML can change without warning. The API also needs basic operational visibility.

## Decision

Expose `/metrics` using `prometheus/client_golang`. Track HTTP request totals and durations, ETL run totals, ETL duration, last successful ETL timestamp, parser issue count, and database row counts.

## Consequences

The server can be observed by a standard Prometheus stack without custom adapters. Metrics are public through the Compose Nginx config for the MVP; production deployments may restrict `/metrics` by network or auth.
