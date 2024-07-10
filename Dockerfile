FROM golang:latest
EXPOSE 80

WORKDIR /app
COPY . .

RUN go build .
ENTRYPOINT ["./watchify-server"]
