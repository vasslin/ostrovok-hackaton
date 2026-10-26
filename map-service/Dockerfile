FROM golang:1.27.1 AS build
WORKDIR /src
COPY . .
WORKDIR /src/map-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/app

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata netcat-openbsd
COPY --from=build /out/service /app/service
USER 10001:10001
EXPOSE 50056
ENTRYPOINT ["/app/service"]
