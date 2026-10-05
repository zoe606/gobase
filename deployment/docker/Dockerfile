# Step 1: Modules caching
FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine AS modules

COPY go.mod go.sum /modules/

WORKDIR /modules

RUN go mod download

# Step 2: Builder
FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine AS builder

ARG TARGET=app
ARG TARGETOS
ARG TARGETARCH

COPY --from=modules /go/pkg /go/pkg
COPY . /app

WORKDIR /app

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /bin/app ./cmd/${TARGET}

# Step 3: Final
FROM scratch

COPY --from=builder /app/config /config
COPY --from=builder /app/migrations /migrations
COPY --from=builder /bin/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app"]
