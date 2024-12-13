FROM golang:1.23
LABEL authors="iskiridomov"

WORKDIR /app
COPY . .

RUN go build -o main .

ENTRYPOINT ["./main"]
