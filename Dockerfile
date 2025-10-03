FROM golang:1.25.1 AS builder

WORKDIR /app
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s"

FROM scratch

COPY --from=builder /app/rockon-validator /bin/rockon-validator

VOLUME /files
WORKDIR /files
ENTRYPOINT [ "/bin/rockon-validator" ]
