.PHONY: local-setup
local-setup:
	@./scripts/local-setup.sh

.PHONY: local-serve
local-serve:
	@go run main.go serve

.PHONY: local-clean
local-clean:
	@rm -f storage/logs/*.log

.PHONY: local-fresh
local-fresh:
	@rm -f storage/app/*.txt
	@rm -f storage/app/*.json
	@rm -f storage/database/*.json
	@rm -f storage/logs/*.log

.PHONY: build
build:
	@GOOS=linux GOARCH=amd64 go build -o p-node

.PHONY: setup
setup:
	@./scripts/setup.sh

.PHONY: update
update:
	@git fetch --all
	@git reset --hard
	@git clean -fd
	@git pull
	@./scripts/setup.sh

.PHONY: info
info:
	@./scripts/info.sh

.PHONY: set-manager
set-manager:
	@./scripts/set-manager.sh "$(URL)" "$(TOKEN)"

.PHONY: fresh
fresh:
	@rm -f storage/app/*.txt
	@rm -f storage/app/*.json
	@rm -f storage/database/*.json
	@rm -f storage/logs/*.log
	@docker compose restart

.PHONY: schedule-server-reboot
schedule-server-reboot:
	@./scripts/schedule-server-reboot.sh
