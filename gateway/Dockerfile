FROM golang:1.27.1 AS build
WORKDIR /src
COPY . .
WORKDIR /src/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gateway ./cmd/app

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/gateway /app/gateway
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
