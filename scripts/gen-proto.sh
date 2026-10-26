#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
protoc \
  -I . \
  --go_out=pkg --go_opt=module=github.com/godnota/pkg \
  --go-grpc_out=pkg --go-grpc_opt=module=github.com/godnota/pkg \
  --go-grpc_opt=require_unimplemented_servers=false \
  api/common/v1/trip.proto \
  api/session/v1/session.proto \
  api/user/v1/user.proto \
  api/order/v1/order.proto \
  api/travel/v1/travel.proto \
  api/ai/v1/ai.proto \
  api/map/v1/map.proto \
  api/route/v1/route.proto \
  api/provider/v1/provider.proto
