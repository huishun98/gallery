FROM golang:1.24-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# No CGO needed
ENV CGO_ENABLED=0
RUN go build -o /out/gallery ./main.go


FROM debian:bookworm-slim

# Runtime deps (no sqlite libs)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    ffmpeg \
  && rm -rf /var/lib/apt/lists/*

# Install cloudflared (official binary)
RUN curl -fsSL \
      https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
      -o /usr/local/bin/cloudflared \
  && chmod +x /usr/local/bin/cloudflared

WORKDIR /app

COPY --from=builder /out/gallery /usr/local/bin/gallery

RUN useradd -m appuser && chown -R appuser /app
USER appuser

EXPOSE 8000

ENTRYPOINT ["gallery"]
