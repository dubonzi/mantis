dev:
	go run -race cmd/*.go --conf=configs/base.yaml

tests: 
	go test -race -v -count=1 ./...

cover:
	go test -race -coverprofile=coverage.out -count=1 ./... && go tool cover -html=coverage.out && rm coverage.out

#Requires gotestsum (https://github.com/gotestyourself/gotestsum)
testsum:
	gotestsum --packages="./..." -- -count=1 -race -cover

testwatch:
	gotestsum --watch --packages="./..." -- -count=1 -race -cover
	