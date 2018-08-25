.PHONY: build clean upload
build:
	go build
	zip -r dotateacher.zip dotateacher

clean:
	if [ -a dotateacher.zip ]; then rm dotateacher.zip; fi

upload: build
	aws lambda update-function-code \
--zip-file=fileb://dotateacher.zip \
--region=eu-central-1 \
--function-name=dotateacher