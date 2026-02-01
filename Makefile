MODULE_PATH := github.com/alexey-y-a/go-userauth-microservices

.PHONY: help
help:
	@echo "Доступные команды:"
	@echo "  make help   - показать это сообщение"
	@echo "  make tidy   - go mod tidy"
	@echo "  make test   - запустить все тесты"

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test ./...
