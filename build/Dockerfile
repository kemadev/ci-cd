ARG ALPINE_MAJOR_MINOR_VERSION=3.21
ARG ALPINE_MAJOR_MINOR_PATCH_VERSION=3.21.3

ARG HADOLINT_VERSION=v2.12.0
ARG ACTIONLINT_VERSION=1.7.7
ARG GITLEAKS_VERSION=v8.24.3
ARG GOLANGCI_LINT_VERSION=v2.1.2
ARG SHELLCHECK_VERSION=v0.10.0
ARG KUBECTL_VERSION=1.32.3
ARG PULUMI_VERSION=3.162.0
ARG GRYPE_VERSION=v0.91.0
ARG SYFT_VERSION=v1.22.0
ARG GORELEASER_VERSION=v2.8.2
ARG GO_VERSION=1.24.2

ARG NODEJS_VERSION=23.11.0
ARG PYTHON_VERSION=3.13.3

FROM --platform=${BUILDPLATFORM} docker.io/hadolint/hadolint:${HADOLINT_VERSION} AS dep-hadolint
FROM --platform=${BUILDPLATFORM} docker.io/rhysd/actionlint:${ACTIONLINT_VERSION} AS dep-actionlint
FROM --platform=${BUILDPLATFORM} docker.io/zricethezav/gitleaks:${GITLEAKS_VERSION} AS dep-gitleaks
FROM --platform=${BUILDPLATFORM} docker.io/golangci/golangci-lint:${GOLANGCI_LINT_VERSION} AS dep-golangci-lint
FROM --platform=${BUILDPLATFORM} docker.io/koalaman/shellcheck:${SHELLCHECK_VERSION} AS dep-shellcheck
FROM --platform=${BUILDPLATFORM} docker.io/bitnami/kubectl:${KUBECTL_VERSION} AS dep-kubectl
FROM --platform=${BUILDPLATFORM} docker.io/pulumi/pulumi-go:${PULUMI_VERSION} AS dep-pulumi
FROM --platform=${BUILDPLATFORM} docker.io/anchore/grype:${GRYPE_VERSION} AS dep-grype
FROM --platform=${BUILDPLATFORM} docker.io/anchore/syft:${SYFT_VERSION} AS dep-syft
FROM --platform=${BUILDPLATFORM} docker.io/goreleaser/goreleaser:${GORELEASER_VERSION} AS dep-goreleaser
FROM --platform=${BUILDPLATFORM} docker.io/golang:${GO_VERSION}-alpine${ALPINE_MAJOR_MINOR_VERSION} AS dep-go
FROM --platform=${BUILDPLATFORM} docker.io/node:${NODEJS_VERSION}-alpine${ALPINE_MAJOR_MINOR_VERSION} AS dep-nodejs
FROM --platform=${BUILDPLATFORM} docker.io/python:${PYTHON_VERSION}-alpine${ALPINE_MAJOR_MINOR_VERSION} AS dep-python

FROM --platform=${BUILDPLATFORM} docker.io/alpine:${ALPINE_MAJOR_MINOR_PATCH_VERSION} AS base-deps

# hadolint ignore=DL3018
RUN apk add --no-cache \
	libstdc++ \
	python3 \
	gcc \
	musl-dev \
	docker \
	buildkit \
	bash \
	curl \
	jq \
	git \
	tar \
	gzip \
	zip

FROM --platform=${BUILDPLATFORM} base-deps AS misc-deps

# Docker buildx ARGs
ARG TARGETOS
ARG TARGETARCH

ARG HELM_VERSION=v3.17.3
ARG RELEASE_PLEASE_VERSION=17.0.0
ARG MARKDOWNLINT_VERSION=0.44.0
ARG SEMGREP_VERSION=1.119.0

# hadolint ignore=DL4006
RUN curl -sSL "https://get.helm.sh/helm-${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz" | tar -xz --strip-components=1 -C /usr/local/bin "${TARGETOS}-${TARGETARCH}/helm"

COPY --link --from=dep-nodejs /usr/local /usr/local

RUN npm --cache=false install --global "release-please@${RELEASE_PLEASE_VERSION}" "markdownlint-cli@${MARKDOWNLINT_VERSION}"

COPY --link --from=dep-python /usr/local/ /usr/local/

RUN pip install --no-cache-dir "semgrep==${SEMGREP_VERSION}"

FROM --platform=${BUILDPLATFORM} dep-nodejs AS nodejs-deps

RUN mkdir /go && \
	rm -rf /usr/local/CHANGELOG.md /usr/local/LICENSE /usr/local/README.md /usr/local/bin/yarn /usr/local/bin/yarnpkg /usr/local/bin/docker-entrypoint.sh

FROM --platform=${BUILDPLATFORM} base-deps AS clean-deps

COPY --link --from=dep-hadolint /bin/hadolint /usr/local/bin/hadolint
COPY --link --from=dep-actionlint /usr/local/bin/actionlint /usr/local/bin/actionlint
COPY --link --from=dep-gitleaks /usr/bin/gitleaks /usr/local/bin/gitleaks
COPY --link --from=dep-golangci-lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint
COPY --link --from=dep-shellcheck /bin/shellcheck /usr/local/bin/shellcheck
COPY --link --from=dep-kubectl /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/kubectl
COPY --link --from=dep-pulumi /pulumi/bin/pulumi /usr/local/bin/pulumi
COPY --link --from=dep-grype /grype /usr/local/bin/grype
COPY --link --from=dep-syft /syft /usr/local/bin/syft
COPY --link --from=dep-goreleaser /usr/bin/goreleaser /usr/local/bin/goreleaser
COPY --link --from=dep-go /usr/local/go/ /usr/local/go/
COPY --link --from=nodejs-deps /usr/local/ /usr/local/
COPY --link --from=dep-python /usr/local/ /usr/local/
COPY --link --from=misc-deps /usr/local/lib/ /usr/local/lib/
COPY --link --from=misc-deps /usr/local/bin/helm /usr/local/bin/helm
COPY --link --from=misc-deps /usr/local/bin/pysemgrep /usr/local/bin/semgrep

RUN ln -s /usr/local/lib/node_modules/release-please/build/src/bin/release-please.js /usr/local/bin/release-please && ln -s /usr/local/lib/node_modules/markdownlint-cli/markdownlint.js /usr/local/bin/markdownlint

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH="${PATH}:${GOROOT}/bin:${GOPATH}/bin"

ARG GIT_USERNAME="github-actions[bot]"
ARG GIT_EMAIL="kemadev+github-actions[bot]@users.noreply.github.com"

RUN git config --global user.name "${GIT_USERNAME}" && git config --global user.email "${GIT_EMAIL}"

ENV SEMGREP_IN_DOCKER=1

ENV PYTHONIOENCODING=utf8 PYTHONUNBUFFERED=1

RUN mkdir -p /src && chown root:root /src && git config --global --add safe.directory /src

WORKDIR /src

RUN chmod 755 /usr/local/go/bin/* && chown root:root /usr/local/go/bin/*

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
	go build -ldflags='-s -w' -o /usr/local/bin/kema-runner ./cmd/main

FROM --platform=${BUILDPLATFORM} clean-deps AS clean

ENV HOME=/home/nobody

RUN mkdir -p ${HOME} && chown nobody:nobody ${HOME}

USER nobody:nobody

FROM clean-deps AS dev

# Docker buildx ARGs
ARG TARGETOS
ARG TARGETARCH

ARG CGO_ENABLED=0
ARG GOAMD64=v3
ARG GOARM64=v8.2

ENV CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM64=${GOARM64} GOAMD64=${GOAMD64}

RUN go install github.com/air-verse/air@latest

ENTRYPOINT ["air", "-c", "config/air/.air.toml"]

FROM clean-deps AS goreleaser

LABEL org.opencontainers.image.authors="kemadev"
LABEL org.opencontainers.image.url="ghcr.io/kemadev/ci-cd"
LABEL org.opencontainers.image.documentation="https://github.com/kemadev/ci-cd"
LABEL org.opencontainers.image.source="https://github.com/kemadev/ci-cd"
LABEL org.opencontainers.image.vendor="kemadev"
LABEL org.opencontainers.image.licenses="MPL-2.0"

COPY {{ vcs_repo }} /usr/local/bin/kema-runner

CMD ["/usr/local/bin/kema-runner"]

FROM --platform=${BUILDPLATFORM} clean-deps AS runner

LABEL org.opencontainers.image.authors="kemadev"
LABEL org.opencontainers.image.url="ghcr.io/kemadev/ci-cd"
LABEL org.opencontainers.image.documentation="https://github.com/kemadev/ci-cd"
LABEL org.opencontainers.image.source="https://github.com/kemadev/ci-cd"
LABEL org.opencontainers.image.vendor="kemadev"
LABEL org.opencontainers.image.licenses="MPL-2.0"

COPY --link --from=build /usr/local/bin/kema-runner /usr/local/bin/kema-runner

ENTRYPOINT ["/usr/local/bin/kema-runner"]
