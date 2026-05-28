# Build stage
FROM golang:1.26.3-alpine AS build

WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Single binary, every entrypoint via subcommand (qp serve:api, qp
# serve:worker, qp serve:cron, qp migrate:up, qp admin:grant, ...).
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/qp \
    ./cmd/qp

# Runtime stage
FROM chainguard/static:latest

COPY --from=build /out/qp /app/qp

USER 65532:65532
EXPOSE 3000

ENTRYPOINT ["/app/qp"]
CMD ["serve:api"]
