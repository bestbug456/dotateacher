.PHONY: build clean upload
build:
	go build
	mv src dotateacher
	zip -r dotateacher.zip dotateacher

clean:
	if [ -a dotateacher.zip ]; then rm dotateacher.zip; fi

upload: build
	./upload.sh