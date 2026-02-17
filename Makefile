info:
	cloc .
run-test:
	go test -cover 
coverage:
	go test -coverprofile=coverage.out 
	go tool cover -html=coverage.out
build-plugin:
	go build -buildmode=plugin -o loglint.so ./plugin