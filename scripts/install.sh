#!/usr/bin/env bash

if ! command -v pip &> /dev/null
then
    echo "Error: pip could not be found. Please install it or add it to your PATH."
    exit 1
fi

if [ -f "requirements.txt" ]; then
    pip install -r requirements.txt
else
    pip install pytest requests ruff playwright pytest-playwright
fi

# Install Playwright browsers (chromium only for faster install)
playwright install chromium

go install github.com/pressly/goose/v3/cmd/goose@latest
