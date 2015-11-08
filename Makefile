build:
	go build main.go
	mkdir -p dist
	rm -rfv dist/*
	mv main dist/plutonium
	cp -R data dist/data
	chmod -R 777 dist/
