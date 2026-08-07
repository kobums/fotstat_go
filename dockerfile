FROM --platform=linux/amd64 golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# .env.yml 은 이미지에 굽지 않는다 — 비밀값이 공개 레지스트리에 노출되므로
# 운영 서버 compose 가 /data/fotstat_go/.env.yml 을 /root/.env.yml 로 마운트한다

# Creating webdata directory in the CURRENT working directory
RUN mkdir -p webdata


ENV APP_MODE=production

CMD ["./main"]