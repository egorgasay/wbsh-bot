FROM golang:1.19

# Set destination for COPY
WORKDIR /app

COPY . .

# Build
RUN go mod download && mkdir /data && mv schedule.db /data && go mod tidy && ls && go build -o /app/main cmd/bot/main.go && pwd && ls && chmod +x /app/main

EXPOSE 80

CMD ["./main"]
VOLUME /data