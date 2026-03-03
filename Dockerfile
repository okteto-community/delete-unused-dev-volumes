FROM golang:1.23.4-bookworm

# Install curl for downloading Okteto CLI at runtime
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

WORKDIR /app
ADD app/ .
RUN go build -o /usr/local/bin/app

# Add entrypoint script that downloads latest Okteto CLI
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]