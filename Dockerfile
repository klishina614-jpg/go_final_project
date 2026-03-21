FROM golang:1.25 AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /todo-app .

FROM ubuntu:latest

WORKDIR /app
COPY --from=build /todo-app .
COPY web/ ./web/

ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db

EXPOSE 7540

CMD ["./todo-app"]
