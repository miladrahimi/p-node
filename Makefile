.PHONY: setup run build info fresh update

setup:
	./third_party/install-xray-mac.sh

run:
	go run main.go start

build:
	go build main.go -o ssm

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
	rm -f storage/database.json
	rm -f storage/xray.json
	docker compose restart

update:
	git pull
	docker compose pull
	docker compose down
	rm -f ./storage/xray.json
	docker compose up -d
