FROM golang:1.26.5 AS build
WORKDIR /src
COPY . .
WORKDIR /src/gateway
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -o /out/gateway ./cmd/app

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/gateway /app/gateway
WORKDIR /app
RUN mkdir -p /app/data && chown 10001:10001 /app/data
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
