FROM golang:1.21

ENV config=docker

WORKDIR /app

COPY . /app

RUN go mod download

RUN go get github.com/githubnemo/CompileDaemon

RUN go install github.com/githubnemo/CompileDaemon

EXPOSE 5000

ENTRYPOINT CompileDaemon --build="go build cmd/email_service/main.go" --command=./main