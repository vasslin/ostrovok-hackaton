FROM alpine:3.20 AS aimock
RUN apk add --no-cache ca-certificates netcat-openbsd
COPY integration-tests/out/aimock /aimock
USER 10001:10001
EXPOSE 50055
ENV GRPC_PORT=50055
ENTRYPOINT ["/aimock"]

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /src/integration-tests
COPY configs/examples /src/configs/examples
COPY integration-tests/out/stage23.test /src/integration-tests/stage23.test
USER 10001:10001
ENTRYPOINT ["/src/integration-tests/stage23.test"]
CMD ["-test.v", "-test.count=1"]
