build:
	go build main.go
	mkdir -p dist
	rm -rfv dist/*
	mv main dist/plutonium
	mkdir -p dist/data
	cp -R data/template dist/data/template
	unzip data/electron-v0.34.3-linux-x64.zip -d dist/data/electron