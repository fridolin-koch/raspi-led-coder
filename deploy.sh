#!/usr/bin/env bash

GOARM=6 GOARCH=arm GOOS=linux go build -o led-coder

scp led-coder $1: