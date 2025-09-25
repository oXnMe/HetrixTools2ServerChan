FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git
COPY . .
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o webhook-app .

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/webhook-app .
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./webhook-app"]
