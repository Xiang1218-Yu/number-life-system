FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/digital-life ./cmd/server
FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/digital-life ./digital-life
COPY config ./config
COPY web ./web
EXPOSE 8080
ENTRYPOINT ["./digital-life"]
