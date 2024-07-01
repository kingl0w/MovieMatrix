FROM golang:1.20

WORKDIR /app

COPY go.mod go.sum ./

RUN apt-get update && apt-get install -y ca-certificates

RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]