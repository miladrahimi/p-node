.PHONY: prepare setup info fresh update

prepare:
	./third_party/install-xray-mac.sh

setup:
	./scripts/setup-updater.sh
	@if [ ! -f ./configs/main.local.json ]; then \
		cp ./configs/main.json ./configs/main.local.json; \
	fi

info:
	@if [ -e "$(CURDIR)/storage/database.json" ]; then \
        printf "IP: "; \
        curl ifconfig.io; \
        printf "DB: "; \
        cat "$(CURDIR)/storage/database.json"; \
        printf "\n"\
    else \
        echo "The app is not ready yet. Please try again..."; \
    fi


fresh:
	rm storage/*.json
	docker compose restart

update: setup
	@echo "$(shell date '+%Y-%m-%d %H:%M:%S') Updating..." >> ./storage/updates.txt
	git pull
	docker compose pull
	docker compose down
	rm -f ./storage/xray.json
	docker compose up -d
	@echo "$(shell date '+%Y-%m-%d %H:%M:%S') Updated." >> ./storage/updates.txt
