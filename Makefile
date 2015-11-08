build:
	go build main.go
	rm -rfv dist/*
	mv main dist/plutonium
	cp -R data dist/data
	cp -R electron dist/electron
	chmod -R 777 dist/
