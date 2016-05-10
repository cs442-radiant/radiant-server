#!/usr/bin/env bash
gofmt -w server/
go install "radiant-server/server"
