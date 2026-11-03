# Stage 1: Build binaries
FROM golang:1.24-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build headless server and CLI client
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/ScopeSentry ./cmd/main && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/scopesentry ./cmd/scopesentry

# Stage 2: Runtime image
FROM alpine:latest

ENV TZ=Asia/Shanghai

RUN apk add --no-cache ca-certificates tzdata curl && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /opt/ScopeSentry

RUN mkdir -p /opt/ScopeSentry/images /opt/ScopeSentry/files /opt/ScopeSentry/uploads

COPY --from=builder /build/ScopeSentry /opt/ScopeSentry/ScopeSentry
COPY --from=builder /build/scopesentry /usr/local/bin/scopesentry
RUN ln -s /opt/ScopeSentry/ScopeSentry /usr/local/bin/ScopeSentry

RUN chmod +x /opt/ScopeSentry/ScopeSentry /usr/local/bin/scopesentry

EXPOSE 8082

ENTRYPOINT ["/opt/ScopeSentry/ScopeSentry"]
