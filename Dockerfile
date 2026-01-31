FROM golang:1.25.6-alpine AS build
#RUN apk add --no-cache curl libstdc++ libgcc alpine-sdk
RUN apk add --no-cache ca-certificates

WORKDIR /app

#RUN curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64-musl -o tailwindcss
#RUN chmod +x tailwindcss
#RUN go install github.com/a-h/templ/cmd/templ@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

#RUN templ generate
#RUN ./tailwindcss -i cmd/web/styles/input.css -o cmd/web/assets/css/output.css
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/main ./cmd/api/main.go

FROM alpine:3.20.1 AS prod
RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=build /app/main /app/main
COPY --from=build /app/ui /app/ui
COPY --from=build /app/config /app/config

ARG PORT=8084
ENV PORT=${PORT}

EXPOSE ${PORT}
CMD ["./main"]
