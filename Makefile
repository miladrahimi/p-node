.PHONY: dev-setup dev-run dev-fresh dev-clean setup info fresh update

dev-setup:
	@./scripts/dev-setup.sh

dev-run:
	@go run main.go start

dev-fresh:
	@rm -f storage/app/*.txt
	@rm -f storage/app/*.json
	@rm -f storage/database/*.json
	@rm -f storage/logs/*.log

dev-clean:
	@rm -f storage/logs/*.log

setup:
	@./scripts/setup.sh

info:
	@./scripts/info.sh

fresh:
	@rm -f storage/app/*.txt
	@rm -f storage/app/*.json
	@rm -f storage/database/*.json
	@rm -f storage/logs/*.log
	@docker compose restart

update: setup
	@git pull
	@./scripts/update.sh
