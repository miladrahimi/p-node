.PHONY: local-setup
local-setup:
	@./scripts/local-setup.sh

.PHONY: local-run
local-run:
	@go run main.go start

.PHONY: local-fresh
local-fresh:
	@rm -f storage/app/*.txt
	@rm -f storage/app/*.json
	@rm -f storage/database/*.json
	@rm -f storage/logs/*.log

.PHONY: local-clean
local-clean:
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
