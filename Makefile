test:
	go test -cover ./...

cover:
	go test -coverprofile cover.out  ./... && go tool cover -html=cover.out

race:
	go test -race -count=100 ./...
