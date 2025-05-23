# syntax=docker/dockerfile:1

## Build
FROM golang:1.24.2-bookworm AS build

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o p-node

## Deploy
FROM ghcr.io/miladrahimi/debian:bookworm-slim

WORKDIR /app

COPY --from=build /app/p-node p-node
COPY --from=build /app/configs/main.defaults.json configs/main.defaults.json
COPY --from=build /app/storage/app/.gitignore storage/app/.gitignore
COPY --from=build /app/storage/database/.gitignore storage/database/.gitignore
COPY --from=build /app/storage/logs/.gitignore storage/logs/.gitignore
COPY --from=build /app/third_party/xray-linux-64/xray third_party/xray-linux-64/xray

EXPOSE 8080

ENTRYPOINT ["./p-node", "start"]
