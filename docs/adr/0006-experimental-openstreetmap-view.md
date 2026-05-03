# ADR 0006: Experimental OpenStreetMap View

Date: 2026-05-03

## Status

Accepted as an experiment.

## Context

Street search across many neighborhoods is useful, but visual orientation may help residents who know an area better than the official neighborhood name. The project does not currently have authoritative neighborhood polygons or street coordinates from RETIM.

## Decision

Add `/map` as an isolated experimental route. It embeds an OpenStreetMap view centered on Arad and pairs it with the real extracted street/neighborhood catalog from the local ETL. Street rows link to the app program page and to OpenStreetMap search for independent verification.

## Consequences

The core app does not depend on map data, geocoding, tile APIs, or third-party JavaScript. If the experiment is not useful, `/map` can be removed without changing ETL, search, print, ICS, or deployment. Future work can add cached geocoding or official GIS polygons if reliable public data becomes available.
