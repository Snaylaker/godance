# syntax=docker/dockerfile:1

FROM golang:1.22

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./dance.db ./
COPY ./html/template/*.html ./html/template/
COPY *.go ./
RUN GOOS=linux go build -o /go-dance
EXPOSE 8080
CMD ["/go-dance"]
