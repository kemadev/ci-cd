# TODO Add renovate tracking
FROM --platform=${BUILDPLATFORM} docker.io/alpine:3.21.3 AS go-deps

SHELL ["/bin/ash", "-o", "pipefail", "-c"]

# Docker buildx ARGs
ARG BUILDPLATFORM

WORKDIR /deps

# hadolint ignore=DL3018
RUN apk add --no-cache curl grep tar gzip zip

ENV BUILDPLATFORM=${BUILDPLATFORM}

RUN --mount=type=bind,source=go.mod,target=go.mod curl -fsSL "https://go.dev/dl/go$(grep -oP 'go\s+\K([0-9.]+)' <go.mod).$(echo ${BUILDPLATFORM} | sed 's|/|-|g').tar.gz" | tar -C /usr/local -xz

# Ensure future GOPATH exists
RUN mkdir -p /root/go

# TODO Add renovate tracking
FROM --platform=${BUILDPLATFORM} docker.io/alpine:3.21.3 AS clean-deps

WORKDIR /deps

COPY --link --from=go-deps /usr/local/go /usr/local/go
COPY --link --from=go-deps /root/go /root/go

ENV PATH="/root/go/bin:/usr/local/go/bin:${PATH}"

# TODO Add renovate tracking
FROM --platform=${BUILDPLATFORM} docker.io/alpine:3.21.3 AS dev

# Docker buildx ARGs
ARG TARGETOS
ARG TARGETARCH

ARG CGO_ENABLED=0
ARG GOAMD64=v3
ARG GOARM64=v8.2

ENV CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM64=${GOARM64} GOAMD64=${GOAMD64}

# Use manually mounted volume for dev, providing hot-reload possibility
WORKDIR /app

COPY --link --from=clean-deps /usr/local/go /usr/local/go

# hadolint ignore=DL3018
RUN apk add --no-cache git sudo jq && adduser -D -u 1000 nonroot && echo "nonroot ALL=(ALL) NOPASSWD: ALL" >/etc/sudoers.d/nonroot && chmod 0440 /etc/sudoers.d/nonroot && chown nonroot:nonroot /usr/local/go

ENV PATH="/home/nonroot/go/bin:/usr/local/go/bin:${PATH}"

USER 1000:1000

WORKDIR /src

COPY . /src

# Build and run
ENTRYPOINT ["sh", "-c", "cp /run/secrets/git_netrc ~/.netrc && go build -o /app ./cmd/${GO_MAIN_FILE_DIR} && /app"]

FROM --platform=${BUILDPLATFORM} clean-deps AS base

# Docker buildx ARGs
ARG TARGETOS
ARG TARGETARCH

ARG CGO_ENABLED=0
ARG GOAMD64=v3
ARG GOARM64=v8.2

ENV CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM64=${GOARM64} GOAMD64=${GOAMD64}

RUN --mount=type=secret,id=gitcreds \
	--mount=type=cache,target=/root/go/pkg \
	--mount=type=bind,source=go.mod,target=go.mod \
	--mount=type=bind,source=go.sum,target=go.sum \
	cp /run/secrets/gitcreds /root/.netrc && \
	go mod download -x

FROM --platform=${BUILDPLATFORM} base AS build

RUN --mount=type=cache,target=/root/go/pkg \
	--mount=type=cache,target=/root/.cache \
	--mount=type=bind,source=.,target=. \
	go build -ldflags='-s -w' -o /usr/local/bin/app ./cmd/main

FROM gcr.io/distroless/static-debian12:nonroot AS goreleaser

LABEL org.opencontainers.image.authors="kemadev"
LABEL org.opencontainers.image.url="ghcr.io/kemadev/ci-cd"
LABEL org.opencontainers.image.documentation="https://github.com/kemadev/ci-cd"
LABEL org.opencontainers.image.source="https://github.com/kemadev/ci-cd"
LABEL org.opencontainers.image.vendor="kemadev"
LABEL org.opencontainers.image.licenses="MPL-2.0"

COPY ci-cd /app

ENTRYPOINT ["/app"]

# hadolint ignore=DL3007
FROM gcr.io/distroless/static-debian12:nonroot AS main

LABEL org.opencontainers.image.authors="kemadev"
LABEL org.opencontainers.image.url="ghcr.io/kemadev/ci-cd"
LABEL org.opencontainers.image.documentation="https://github.com/kemadev/ci-cd"
LABEL org.opencontainers.image.source="https://github.com/kemadev/ci-cd"
LABEL org.opencontainers.image.vendor="kemadev"
LABEL org.opencontainers.image.licenses="MPL-2.0"

COPY --link --from=build /usr/local/bin/app /app

ENTRYPOINT ["/app"]
