test:
	go test -cover ./...

cover:
	go test -coverprofile cover.out  ./... && go tool cover -html=cover.out

race:
	go test -race -cpu=1,9,55,99 -count=100 -failfast ./...
