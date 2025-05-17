FROM golang:bookworm

WORKDIR /surveysvc

COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download

COPY ./ ./


CMD [ "go", "run", "./cmd/server/main.go" ]