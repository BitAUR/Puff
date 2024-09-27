FROM golang:1.23-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct 

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o puff .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/puff .

RUN mkdir -p /app/data

COPY ./templates /app/templates

EXPOSE 8080

# 运行应用
CMD ["./puff"]