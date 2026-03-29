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
    && rm -rf /var/lib/apt/lists/*
# install bun
RUN curl -fsSL https://bun.sh/install | bash

RUN { \
    echo "alias ls='ls --color=auto'"; \
    echo "alias ll='ls -la --color=auto'"; \
    echo "alias grep='grep --color=auto'"; \
    echo "export PS1='\\[\\033[1;32m\\][dev-container]\\[\\033[0m\\] \\w \\$ '"; \
} >> ~/.bashrc
RUN
RUN mkdir /root/warchest
WORKDIR /root/warchest
