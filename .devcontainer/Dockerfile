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
    # needed for db migrations
    && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
    && go install github.com/pressly/goose/v3/cmd/goose@latest \
    # needed for proto generation
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    # needed for grpc
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    && curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin

# set-up bashrc to have basic colours
RUN { \
    echo '# Function to get the current git branch'; \
    echo 'parse_git_branch() {'; \
    echo '  git branch 2> /dev/null | sed -e "/^[^*]/d" -e "s/* \(.*\)/ (\1)/"'; \
    echo '}'; \
    echo ''; \
    echo '# Set custom colored prompt: user@[devcontainer] path (git branch)'; \
    echo 'export PS1="\[\033[0;32m\]\u\[\033[0m\]@\[\033[1;32m\][devcontainer]\[\033[0m\] \[\033[0;34m\]\w\[\033[0m\]\[\033[0;33m\]\$(parse_git_branch)\[\033[0m\] $ "'; \
    echo '# Enable colors for ls'; \
    echo 'alias ls="ls --color=auto"'; \
    echo 'alias ll="ls -alF --color=auto"'; \
    echo 'alias la="ls -A --color=auto"'; \
    echo ''; \
    echo '# Enable colored auto-completion (for when you hit Tab)'; \
    echo 'bind "set colored-stats on"'; \
    echo 'bind "set colored-completion-prefix on"'; \
} >> ~/.bashrc
# marks git as safe so we can see branches in the .bashrc
RUN git config --global --add safe.directory '*'

# make warchest folder
RUN mkdir /root/warchest
WORKDIR /root/warchest
