FROM golang:1.19

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN swag init

RUN go build -o app .

CMD ["./app"]
