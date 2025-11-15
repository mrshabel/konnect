FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go

# run only executable in final image
FROM alpine

WORKDIR /app
COPY --from=builder main ./main
CMD ["./main.go"]
