#!/bin/bash

sam package --template-file template.yml --output-template-file package.yml --s3-bucket quantumsense-sam
sam deploy --template-file package.yml --capabilities CAPABILITY_IAM --stack-name lambda-eks-test
