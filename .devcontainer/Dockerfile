FROM ubuntu:26.04

ENV DEBIAN_FRONTEND=noninteractive
# install linux bash commands which wil be used
RUN apt-get update && apt-get install -y \
    sudo \
    curl \
    git \
    vim \ 
    unzip \
    ripgrep \
    golang \
    && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL https://bun.com/install | bash

# default go tooling
ENV PATH="/root/go/bin:${PATH}"
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
    && go install github.com/pressly/goose/v3/cmd/goose@latest \
    && curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin

# set-up bashrc to have basic colours
RUN { \
    echo "alias ls='ls --color=auto'"; \
    echo "alias ll='ls -la --color=auto'"; \
    echo "alias grep='grep --color=auto'"; \
    echo "export PS1='\\[\\033[1;32m\\][dev-container]\\[\\033[0m\\] \\w \\$ '"; \
} >> ~/.bashrc

# make warchest folder
RUN mkdir /root/warchest
WORKDIR /root/warchest
