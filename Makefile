dev:
	go run -race cmd/*.go --conf=configs/base.yaml

tests: 
	go test -race -v -count=1 -shuffle=on ./...

cover:
	go test -race -coverprofile=coverage.out -count=1 -shuffle=on ./... && go tool cover -html=coverage.out && rm coverage.out

cover-ci:
	go test -race -coverprofile=coverage.out -count=1 -shuffle=on ./...

#Requires gotestsum (https://github.com/gotestyourself/gotestsum)
testsum:
	gotestsum --packages="./..." -- -count=1 -race -cover -shuffle=on

testwatch:
	gotestsum --watch --packages="./..." -- -count=1 -race -cover
	