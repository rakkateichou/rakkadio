FROM golang:1.21.5 as build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /rakkadio


FROM alpine:3.19.1 as prod-stage

WORKDIR /root

COPY --from=build-stage /rakkadio ./rakkadio

EXPOSE 8080

CMD ["./rakkadio"]
