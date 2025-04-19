.PHONY: dev-setup
dev-setup:
	@./scripts/dev-setup.sh

.PHONY: dev-run
dev-run:
	@go run main.go start

.PHONY: dev-fresh
dev-fresh:
	@rm -f storage/app/*.txt
	@rm -f storage/app/*.json
	@rm -f storage/database/*.json
	@rm -f storage/logs/*.log

.PHONY: dev-clean
dev-clean:
	@rm -f storage/logs/*.log

.PHONY: schedule-reboot
schedule-reboot:
	@./scripts/schedule-reboot.sh

.PHONY: build
build:
	@GOOS=linux GOARCH=amd64 go build -o p-node

.PHONY: setup
setup:
	@./scripts/setup.sh

.PHONY: update
update:
	@git reset --hard HEAD^
	@git pull
	@./scripts/setup.sh
	@./scripts/update.sh

.PHONY: info
info:
	@./scripts/info.sh

.PHONY: fresh
fresh:
	@rm -f storage/app/*.txt
	@rm -f storage/app/*.json
	@rm -f storage/database/*.json
	@rm -f storage/logs/*.log
	@docker compose restart
