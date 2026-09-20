
FROM golang:1-alpine3.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o canvas-go .

FROM scratch

COPY --from=builder /app/canvas-go /canvas-go

ENTRYPOINT [ "/canvas-go" ]
