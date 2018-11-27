#!/bin/bash

GOOS=linux go build handler.go
mv handler bin
