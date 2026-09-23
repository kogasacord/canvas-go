
FROM golang:1-alpine3.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o canvas-go .

FROM scratch

COPY --from=builder /app/canvas-go /canvas-go
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/images/quote/qgradient.png /images/quote/qgradient.png

ENTRYPOINT [ "/canvas-go" ]
