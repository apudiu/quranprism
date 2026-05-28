# Build stage
FROM golang:1.26.3-alpine AS build

WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build all three binaries
ARG TARGET=api
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/binary \
    ./cmd/${TARGET}

# Runtime stage
FROM chainguard/static:latest

COPY --from=build /out/binary /app/binary

USER 65532:65532
EXPOSE 3000

ENTRYPOINT ["/app/binary"]
