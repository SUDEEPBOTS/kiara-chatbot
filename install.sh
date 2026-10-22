#!/usr/bin/env bash
# ==============================================================================
# Copyright (c) 2026 SUDEEPBOTS (Sudeep Sharma). All rights reserved.
#
# Project: Kiara Chatbot - High-Performance Telegram AI Assistant
# Repository: https://github.com/SUDEEPBOTS/kiara-chatbot
# License: MIT License
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
# ==============================================================================

set -e

RED="\033[1;31m"
GREEN="\033[1;32m"
YELLOW="\033[1;33m"
BLUE="\033[1;34m"
CYAN="\033[1;36m"
BOLD="\033[1m"
RESET="\033[0m"

INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_NAME="kiara"
SERVICE_NAME="kiara.service"

print_banner() {
    clear 2>/dev/null || true
    echo -e "${CYAN}${BOLD}"
    cat << "EOF"
  _  ___                  ____ _           _   _           _   
 | |/ (_) __ _ _ __ __ _ / ___| |__   __ _| |_| |__   ___ | |_ 
 | ' /| |/ _` | '__/ _` | |   | '_ \ / _` | __| '_ \ / _ \| __|
 | . \| | (_| | | | (_| | |___| | | | (_| | |_| |_) | (_) | |_ 
 |_|\_\_|\__,_|_|  \__,_|\____|_| |_|\__,_|\__|_.__/ \___/ \__|
EOF
    echo -e "${RESET}"
    echo -e "${BLUE}================================================================${RESET}"
    echo -e "${BOLD}       Kiara Telegram AI Chatbot - Automated Installer          ${RESET}"
    echo -e "${BLUE}================================================================${RESET}\n"
}

print_step() {
    echo -e "\n${BLUE}${BOLD}▶ $1${RESET}"
}

print_success() {
    echo -e "${GREEN}✔ $1${RESET}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${RESET}"
}

print_error() {
    echo -e "${RED}✖ $1${RESET}"
    exit 1
}

check_dependencies() {
    print_step "Checking system dependencies..."

    if ! command -v curl >/dev/null 2>&1; then
        print_warning "curl is missing. Installing..."
        sudo apt-get update && sudo apt-get install -y curl || sudo yum install -y curl
    fi

    if ! command -v ffmpeg >/dev/null 2>&1; then
        print_warning "FFmpeg is missing (required for video stickers). Installing..."
        sudo apt-get update && sudo apt-get install -y ffmpeg || sudo yum install -y ffmpeg
    fi
    print_success "FFmpeg is installed ($(ffmpeg -version 2>&1 | head -n 1 | awk '{print $3}'))"

    if ! command -v go >/dev/null 2>&1; then
        print_warning "Golang is missing. Please install Go 1.23+ before proceeding."
        print_error "Install Go via: sudo snap install go --classic  OR  https://go.dev/dl/"
    fi
    print_success "Golang is installed ($(go version | awk '{print $3}'))"
}

setup_env() {
    print_step "Configuring environment (.env)..."
    cd "$INSTALL_DIR"

    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            print_success "Created .env from .env.example"
        else
            touch .env
        fi
    else
        print_success "Existing .env file found."
    fi

    if grep -q "BOT_TOKEN=$" .env 2>/dev/null || grep -q "BOT_TOKEN=your_" .env 2>/dev/null; then
        echo -e "\n${YELLOW}${BOLD}Telegram Bot Token is not configured!${RESET}"
        read -r -p "Enter your Telegram Bot Token from @BotFather: " user_token
        if [ -n "$user_token" ]; then
            sed -i "s/^BOT_TOKEN=.*/BOT_TOKEN=${user_token}/" .env
            print_success "BOT_TOKEN updated in .env"
        fi
    fi
}

build_binary() {
    print_step "Compiling Kiara Chatbot binary..."
    cd "$INSTALL_DIR"

    mkdir -p data
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BIN_NAME" cmd/bot/main.go

    if [ -f "$BIN_NAME" ]; then
        chmod +x "$BIN_NAME"
        print_success "Binary successfully built: $INSTALL_DIR/$BIN_NAME ($(du -h "$BIN_NAME" | cut -f1))"
    else
        print_error "Build failed! Please check Go compiler errors."
    fi
}

setup_systemd() {
    print_step "Systemd Service Setup (Optional)..."

    read -r -p "Do you want to install Kiara as a 24/7 background systemd service? (y/N): " install_svc
    if [[ "$install_svc" =~ ^[Yy]$ ]]; then
        USER_NAME="$(whoami)"
        SERVICE_FILE="/tmp/${SERVICE_NAME}"

        cat << EOF > "$SERVICE_FILE"
[Unit]
Description=Kiara Telegram AI Chatbot Service
After=network.target

[Service]
Type=simple
User=${USER_NAME}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${BIN_NAME}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

        sudo mv "$SERVICE_FILE" "/etc/systemd/system/${SERVICE_NAME}"
        sudo systemctl daemon-reload
        sudo systemctl enable "${SERVICE_NAME}"
        sudo systemctl restart "${SERVICE_NAME}"

        print_success "Systemd service installed and started!"
        echo -e "${CYAN}Commands to manage your service:${RESET}"
        echo -e "  • Check status:  ${BOLD}sudo systemctl status ${SERVICE_NAME}${RESET}"
        echo -e "  • View logs:     ${BOLD}journalctl -u ${SERVICE_NAME} -f${RESET}"
        echo -e "  • Restart bot:   ${BOLD}sudo systemctl restart ${SERVICE_NAME}${RESET}"
        echo -e "  • Stop bot:      ${BOLD}sudo systemctl stop ${SERVICE_NAME}${RESET}"
    else
        print_success "Skipped systemd service installation."
    fi
}

main() {
    print_banner
    check_dependencies
    setup_env
    build_binary
    setup_systemd

    echo -e "\n${GREEN}${BOLD}================================================================${RESET}"
    echo -e "${GREEN}${BOLD}       Kiara Chatbot Installation Completed Successfully!       ${RESET}"
    echo -e "${GREEN}${BOLD}================================================================${RESET}"
    echo -e "\nTo start the bot manually:"
    echo -e "  ${BOLD}cd ${INSTALL_DIR} && ./${BIN_NAME}${RESET}\n"
}

main "$@"
