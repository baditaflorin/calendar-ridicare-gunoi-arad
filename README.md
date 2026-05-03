# Gunoi Arad

Calendar simplu, cautabil si printabil pentru programul de ridicare a deseurilor din Municipiul Arad.

MVP-ul descarca sursele publice RETIM, pastreaza snapshot-uri brute cu hash, parseaza tabelele oficiale in SQLite si serveste:

- cautare fuzzy dupa strada si cartier;
- autocomplete cu nume reale extrase din site;
- evenimente recurente si date exacte pentru fractii;
- export ICS;
- pagina printabila A4;
- pagina experimentala `/map` pentru orientare vizuala;
- `/metrics` pentru Prometheus;
- Docker Compose cu Nginx pe portul `26453`.

## Dezvoltare Locala

```bash
go test ./...
go run ./cmd/gunoiarad etl --db .smoke/dev.db --raw-dir .smoke/raw
go run ./cmd/gunoiarad serve --addr :18080 --db .smoke/dev.db --raw-dir .smoke/raw --public-base-url http://localhost:18080
```

Deschide apoi `http://localhost:18080`.

## GitHub Pages Static

Pentru public, varianta recomandata este frontend static + JSON comis in repo. Rulezi ETL local, exporti fisierele in `docs/`, apoi configurezi GitHub Pages sa serveasca branch-ul `main` din folderul `/docs`.

```bash
scripts/export-static.sh
git add docs
git commit -m "data: refresh static calendar export"
git push
```

Frontend-ul static citeste `docs/data/catalog.json` si `docs/data/places/*.json`. Nu are nevoie de backend live pentru cautare, calendar, print A4 sau export ICS.

## Hook-uri Locale

Nu exista GitHub Actions in repo. Pentru verificari locale:

```bash
scripts/install-hooks.sh
scripts/check.sh
scripts/smoke.sh
```

`pre-commit` ruleaza format, vet, teste si validare Compose. `pre-push` ruleaza smoke test cu ETL live impotriva RETIM.

## Docker Compose Server

```bash
docker compose pull
docker compose up -d
```

Nginx expune aplicatia pe `26453`, iar aplicatia asculta intern pe `8080`.

## Build amd64 Pentru GHCR

```bash
docker login ghcr.io
PUSH=true IMAGE=ghcr.io/baditaflorin/calendar-ridicare-gunoi-arad TAG=$(git rev-parse --short HEAD) scripts/build-amd64.sh
```

Pe server:

```bash
docker compose pull
docker compose up -d
```

## Date Si Secrete

Nu comite `.env`, baze SQLite, snapshot-uri din `data/raw`, fisiere din `.smoke` sau orice credential. Sursele publice sunt documentate in `docs/sources.md`, iar deciziile arhitecturale sunt in `docs/adr/`.
