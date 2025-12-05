#!/bin/bash
# Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y nodejs build-essential

# Go
wget https://go.dev/dl/go1.22.7.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.7.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc

# PostgreSQL client
apt-get install -y postgresql-client

# Redis client
apt-get install -y redis-tools

# Docker CLI (для docker-compose)
apt-get install -y docker-compose

# TON SDK (FunC)
# Здесь вставьте актуальные команды установки для Linux
