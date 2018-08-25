.PHONY: build clean upload
build:
	go build
	zip -r dotateacher.zip dotateacher

clean:
	if [ -a dotateacher.zip ]; then rm dotateacher.zip; fi

upload: build
	./upload.sh