FROM golang:1.25.6

ENV GOFLAGS="-p=2"
ENV GOMAXPROCS=2

WORKDIR /app

COPY go.mod go.sum ./

# Install our third-party application for hot-reloading capability.
RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]
RUN ["go", "install", "github.com/githubnemo/CompileDaemon"]

RUN go mod download

# Copy application data into image
#COPY . .

ENTRYPOINT CompileDaemon -polling -log-prefix=false -build="go build -o ./.build/docker_spin ./cmd/server/main.go" -command="./.build/docker_spin" -directory="./"
