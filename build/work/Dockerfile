FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy only the source directories needed
COPY core/ core/
COPY platform/operator/ platform/operator/

# Build operator
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Version=${VERSION}" \
    -o /operator \
    ./platform/operator

# Runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /operator /usr/local/bin/operator

EXPOSE 8080

ENTRYPOINT ["operator"]
