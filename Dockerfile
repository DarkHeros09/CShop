# Build stage
FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN echo "appuser:x:10001:10001:App User:/:/sbin/nologin" > /etc/minimal-passwd
RUN apk add --no-cache upx
# Accept build arguments for architecture and OS
ARG GOARCH
ARG GOOS
ENV CGO_ENABLED=0

RUN go mod tidy
RUN go build -ldflags="-s -w" -o main main.go && \
upx --best --lzma main
# && \
#wget -O wait-for.sh https://github.com/eficode/wait-for/releases/download/v2.2.4/wait-for && \
#chmod +x wait-for.sh

# Copy stage
FROM scratch AS copier
WORKDIR /app
COPY --from=builder /app/main .
#COPY --from=builder /app/wait-for.sh .
COPY --from=builder /etc/minimal-passwd /etc/minimal-passwd
# COPY .env.test .

# Run stage
FROM scratch
WORKDIR /app
ENV TZ=Africa/Tripoli
COPY --from=copier /app .
COPY --from=copier /etc/minimal-passwd /etc/passwd
USER appuser
EXPOSE 8080
CMD [ "/app/main" ]