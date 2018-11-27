#!/bin/bash

GOOS=linux go build github.com/kubernetes-sigs/aws-iam-authenticator/cmd/aws-iam-authenticator
mv aws-iam-authenticator bin
