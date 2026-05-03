# ADR 0004: Docker Compose With Nginx Edge

Date: 2026-05-03

## Status

Accepted.

## Context

The target server should run the app with Docker Compose, expose port `26453`, and pull an amd64 image from GitHub Container Registry.

## Decision

Ship a multi-stage Dockerfile for the Go app, a Docker Compose stack with an Nginx front service on port `26453`, and a script that builds and pushes linux/amd64 images with Docker Buildx.

## Consequences

The runtime boundary is clear: Nginx owns public HTTP and the app listens internally. The app image can be built on Apple Silicon and deployed on amd64 servers.
