FROM golang:alpine as builder
LABEL maintainer="Edi Prakoso <ediprakoso@gmail.com>"
RUN apk update && apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/handler/ui ./handler/ui
COPY --from=builder /app/handler/assets ./handler/assets
COPY --from=builder /app/.env .

EXPOSE 8080

#CMD ["./main"]
#CMD ["./main","-mode=cron"]
CMD ./main && ./main -mode=cron
