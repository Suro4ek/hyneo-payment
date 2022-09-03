FROM golang:1.18.1-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./
COPY config.yml ./
RUN go build cmd/main.go

CMD ["./main"]

FROM alpine:3.15.4

WORKDIR /
COPY --from=build /app/config.yml ./
COPY --from=build /app/main /main

EXPOSE 8080

ENTRYPOINT ["/main"]