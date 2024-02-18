.PHONY: prepare setup info fresh update

prepare:
	./third_party/install-xray-mac.sh

setup:
	@./scripts/setup-updater.sh
	@if [ ! -f ./configs/main.local.json ]; then \
		cp ./configs/main.json ./configs/main.local.json; \
	fi

info:
	@./scripts/info.sh

fresh:
	rm storage/*.json
	docker compose restart

update: setup
	@echo "$(shell date '+%Y-%m-%d %H:%M:%S') Updating..." >> ./storage/updates.txt
	git pull
	docker compose pull
	docker compose down
	docker compose up -d
	@echo "$(shell date '+%Y-%m-%d %H:%M:%S') Updated." >> ./storage/updates.txt
