FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.22.2
RUN apk add --no-cache ca-certificates
COPY --from=build /server /server
WORKDIR /app
COPY web ./web
USER 65532:65532
EXPOSE 8080 9090
ENTRYPOINT ["/server"]
