#!/bin/bash
aws lambda update-function-code \
--zip-file=fileb://dotateacher.zip \
--region=eu-central-1 \
--function-name=dotateacher