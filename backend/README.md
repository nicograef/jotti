# jotti backend

## development

```shell
go mod tidy &&
golangci-lint run &&
goimports -w . &&
go vet ./... &&
go test -tags=unit -v -race ./... &&
go build -v ./...
```