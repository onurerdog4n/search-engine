FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o mock-api main.go
EXPOSE 8081
CMD ["./mock-api"]
