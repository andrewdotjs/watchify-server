FROM golang:latest
EXPOSE 8080

WORKDIR /app
COPY . .

RUN go build .
ENTRYPOINT ["./watchify-server"]
