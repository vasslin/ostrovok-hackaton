FROM golang:1.26.5 AS build
WORKDIR /src
COPY . .
WORKDIR /src/provider-service
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/app

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata netcat-openbsd
WORKDIR /app
COPY --from=build /out/service /app/service
COPY --from=build /src/configs /app/configs
ENV PROVIDER_CONFIG_FILE=/app/configs/providers.json
USER 10001:10001
EXPOSE 50058
ENTRYPOINT ["/app/service"]
