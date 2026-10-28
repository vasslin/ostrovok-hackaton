FROM golang:1.26.5 AS build
WORKDIR /src
COPY . .
WORKDIR /src/user-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/app

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata netcat-openbsd
COPY --from=build /out/service /app/service
USER 10001:10001
EXPOSE 50053
ENTRYPOINT ["/app/service"]
