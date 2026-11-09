# Build stage
FROM golang:alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /desafio ./cmd/game

# Final runtime image
FROM alpine:latest

# Instala ferramentas essenciais para a investigação de DevOps
RUN apk add --no-cache \
    bash \
    curl \
    git \
    docker-cli \
    coreutils \
    findutils \
    ncurses \
    tzdata

WORKDIR /workspace

# Copia o binário compilado
COPY --from=builder /desafio /usr/local/bin/desafio

# Configura prompt amigável e ambientado no terminal
RUN echo 'export PS1="\[\033[01;32m\][EAS-58-INVESTIGADOR:\w]\$ \[\033[00m\]"' >> /root/.bashrc && \
    echo 'alias ll="ls -la"' >> /root/.bashrc

ENV WORKSPACE_DIR=/workspace
ENV PORT=8058
ENV TERM=xterm-256color

EXPOSE 8058

CMD ["/usr/local/bin/desafio", "--workspace=/workspace", "--port=8058"]
