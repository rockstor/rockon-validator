FROM golang:1.20 AS builder

WORKDIR /app
COPY . .

RUN go mod download

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"

FROM scratch

COPY --from=builder /app/rockon-validator /bin/rockon-validator

VOLUME /files
WORKDIR /files
ENTRYPOINT [ "/bin/rockon-validator" ]
