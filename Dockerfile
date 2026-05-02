FROM golang:1.25.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY pkg/ ./pkg/
COPY web/ ./web/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server_app .

FROM ubuntu:latest AS final

WORKDIR /app



# Copy application binary
COPY --from=builder /app/server_app .

# Copy static web files
COPY --from=builder /app/web ./web

ENV TODO_PORT=7540
ENV TODO_DBFILE=scheduler.db
ENV TODO_PASSWORD=12345

EXPOSE 7540

CMD ["/app/server_app"]