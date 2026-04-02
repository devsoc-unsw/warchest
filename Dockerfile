FROM ubuntu:24.04

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
    bun \
    && rm -rf /var/lib/apt/lists/*

# default go tooling
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# set-up bashrc to have basic colours
RUN { \
    echo "alias ls='ls --color=auto'"; \
    echo "alias ll='ls -la --color=auto'"; \
    echo "alias grep='grep --color=auto'"; \
    echo "export PS1='\\[\\033[1;32m\\][dev-container]\\[\\033[0m\\] \\w \\$ '"; \
} >> ~/.bashrc

# install Dingo
RUN git clone https://github.com/MadAppGang/dingo.git /tmp/dingo \
    && cd /tmp/dingo \
    && go build -o /usr/local/bin/dingo ./cmd/dingo \
    && rm -rf /tmp/dingo

# make warchest folder
RUN mkdir /root/warchest
WORKDIR /root/warchest
