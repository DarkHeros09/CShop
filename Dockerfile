# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -tags=go_json,nomsgpack -o main main.go && \
wget -O wait-for.sh https://github.com/eficode/wait-for/releases/download/v2.2.3/wait-for && \
chmod +x wait-for.sh

# Copy stage
FROM scratch AS copier
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/wait-for.sh .
COPY app.env .

# Run stage
FROM alpine:latest
WORKDIR /app
ENV TZ=Africa/Tripoli
COPY --from=copier /app .
EXPOSE 8080
CMD [ "/app/main" ]