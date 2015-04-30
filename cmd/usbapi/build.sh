#!/bin/bash
go build -ldflags "-X main.Version 0.2.3 -X main.Commit $(git rev-parse HEAD)" main.go