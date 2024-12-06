# 共通のベースイメージ
ARG GO_VERSION=1.23.3

# 開発用ステージ
FROM golang:${GO_VERSION}-bullseye AS dev

RUN apt-get update && apt-get install -y tzdata
ENV TZ=Asia/Tokyo

RUN mkdir -p /go/github.com/keitatwr/
WORKDIR /go/github.com/keitatwr/todo-app

# go.mod, go.sumをコピー
COPY go.mod go.sum ./

# ライブラリのダウンロード
RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    go mod download -x

# ソースコードをコピー
COPY . .

# airをインストール
RUN go install github.com/air-verse/air@latest

CMD ["air", "-c", ".air.toml"]

# テスト用ステージ
FROM golang:${GO_VERSION}-bullseye AS test

RUN apt-get update && apt-get install -y tzdata
ENV TZ=Asia/Tokyo

WORKDIR /go/github.com/keitatwr/todo-app

# go.mod, go.sumをコピー
COPY go.mod go.sum ./

# ライブラリのダウンロード
RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    go mod download -x

# ソースコードをコピー
COPY . .

# テストの実行
CMD ["go", "test", "./..."]

# ビルド用ステージ
FROM golang:${GO_VERSION}-bullseye AS build

RUN apt-get update && apt-get install -y tzdata
ENV TZ=Asia/Tokyo

WORKDIR /app

# ソースコードをコピー
COPY . .

# 本番用にビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/main.go

# 本番用ステージ
FROM debian:bullseye-slim AS prod

RUN apt-get update && apt-get install -y tzdata
ENV TZ=Asia/Tokyo

WORKDIR /app

# ビルドされたバイナリをコピー
COPY --from=build /app/server /app/server
COPY --from=build /app/.env /app/.env

CMD ["./server"]