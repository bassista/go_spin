FROM golang:1.25.6

ENV GOFLAGS="-p=2"
ENV GOMAXPROCS=2

WORKDIR /app

COPY go.mod go.sum ./

# Install our third-party application for hot-reloading capability.
RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]
RUN ["go", "install", "github.com/githubnemo/CompileDaemon"]

RUN go mod download

# Default UID/GID for the runtime user (overridable at `docker run -e UID=... -e GID=...`)
ENV UID=1000
ENV GID=1000

# Install gosu for dropping root privileges in Debian-based golang image
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends wget ca-certificates; \
	dpkgArch="$(dpkg --print-architecture | awk -F- '{ print $NF }')"; \
	wget -O /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/1.14/gosu-$dpkgArch"; \
	chmod +x /usr/local/bin/gosu; \
	/usr/local/bin/gosu --version; \
	rm -rf /var/lib/apt/lists/*

# Copy application data into image (left commented for dev usage)
#COPY . .

# Copy entrypoint that ensures user/group exist and runs the process as that user
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

# Default command for hot-reload development
CMD ["CompileDaemon","-polling","-log-prefix=false","-build=go build -o ./.build/docker_spin ./cmd/server/main.go","-command=./.build/docker_spin","-directory=./"]
