build:
	vgo build
	vgo vet

test:
	go test -cover
	golint

coverage:
	go test -coverprofile=coverage.out -test.outputdir .
	go tool cover -html=coverage.out

clean:
	vgo clean

fmt:
	gofmt -w *.go
