#!/bin/bash

./build-handler.sh
sam local invoke --no-event LambdaEksTestFunction
