# syntax=docker/dockerfile:1

# =============================================================================
# Dockerfile multi-stage.
#
# Konsep kunci: image yang dikirim ke server TIDAK berisi compiler Go.
# Ia hanya berisi satu file binary. Hasilnya image ~15 MB, bukan ~250 MB.
#
#   Stage 1 (build)   : punya compiler Go, dipakai untuk mengompilasi
#   Stage 2 (runtime) : hanya menerima HASIL dari stage 1
#
# Kamu sudah paham Docker dari dockercase, jadi bagian ini akan terasa
# familiar. Yang baru: pipeline nanti yang menjalankan `docker build` ini.
# =============================================================================

# --- Stage 1: BUILD -----------------------------------------------------------
FROM golang:1.22-alpine AS build

WORKDIR /src

# go.mod disalin lebih dulu, baru source code. Kenapa?
# Docker menyimpan cache per-layer. Selama go.mod tidak berubah, layer
# `go mod download` tidak dijalankan ulang. Build jadi jauh lebih cepat.
COPY go.mod ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 -> binary statis, bisa jalan tanpa library C
# -s -w         -> buang simbol & info debug, ukuran mengecil
# -trimpath     -> path folder laptop tidak ikut tertanam di binary
#
# ARG diisi oleh pipeline lewat `--build-arg`. Inilah cara pipeline
# menstempel versi ke dalam binary yang nanti jalan di server.
ARG VERSION=dev
ARG COMMIT=none

RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
        -o /out/keranjang \
        ./cmd/keranjang

# --- Stage 2: RUNTIME ---------------------------------------------------------
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates && \
    adduser -D -u 10001 appuser

# Hanya SATU file yang disalin dari stage build. Source code, compiler,
# dan seluruh isi /src tidak ikut terbawa.
COPY --from=build /out/keranjang /usr/local/bin/keranjang

# Jangan jalan sebagai root. Ini temuan paling sering di security scan.
USER appuser

EXPOSE 8080

# HEALTHCHECK memberi tahu Docker/orchestrator kapan container benar-benar siap.
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/keranjang"]
