#!/bin/bash
aws lambda update-function-code \
--zip-file=fileb://code.zip \
--region=eu-central-1 \
--function-name=dotateacher