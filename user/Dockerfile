FROM golang:1.21

ENV config=docker

WORKDIR /app

COPY . /app

RUN go mode download

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon --build="go build cmd/email_service/main.go" --command=./main