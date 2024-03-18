# syntax=docker/dockerfile:1

## Build
FROM ghcr.io/miladrahimi/golang:1.22.1-bookworm AS build

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o xray-node

## Deploy
FROM ghcr.io/miladrahimi/debian:bookworm-slim

WORKDIR /app

COPY --from=build /app/xray-node xray-node
COPY --from=build /app/configs/main.json configs/main.json
COPY --from=build /app/storage/app/.gitignore storage/app/.gitignore
COPY --from=build /app/storage/database/.gitignore storage/database/.gitignore
COPY --from=build /app/storage/logs/.gitignore storage/logs/.gitignore
COPY --from=build /app/third_party/xray-linux-64/xray third_party/xray-linux-64/xray

EXPOSE 8080

ENTRYPOINT ["./xray-node", "start"]
