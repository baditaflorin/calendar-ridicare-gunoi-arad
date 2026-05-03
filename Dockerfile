# syntax=docker/dockerfile:1.7
FROM --platform=$TARGETPLATFORM golang:1.26-bookworm AS build

WORKDIR /src
RUN apt-get update \
  && apt-get install -y --no-install-recommends gcc libc6-dev ca-certificates \
  && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/gunoiarad ./cmd/gunoiarad

FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates curl tzdata \
  && rm -rf /var/lib/apt/lists/* \
  && useradd --create-home --uid 10001 --shell /usr/sbin/nologin appuser \
  && mkdir -p /data/raw /data/db \
  && chown -R appuser:appuser /data

COPY --from=build /out/gunoiarad /usr/local/bin/gunoiarad

ENV GUNOI_HTTP_ADDR=:8080 \
    GUNOI_PUBLIC_BASE_URL=http://localhost:26453 \
    GUNOI_DB_PATH=/data/db/gunoiarad.db \
    GUNOI_RAW_DIR=/data/raw \
    GUNOI_REFRESH_INTERVAL=6h \
    GUNOI_BOOTSTRAP_ETL=true \
    TZ=Europe/Bucharest

USER appuser
VOLUME ["/data"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 CMD curl -fsS http://127.0.0.1:8080/readyz >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/gunoiarad"]
CMD ["serve"]
