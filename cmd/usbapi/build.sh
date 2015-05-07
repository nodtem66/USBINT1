#!/bin/bash
go build -o usbapi -v -ldflags "-X main.Version 0.2.3 -X main.Commit $(git rev-parse HEAD)" main.go 