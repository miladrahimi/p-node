.PHONY: setup run build info fresh update

setup:
	./third_party/install-xray-mac.sh

run:
	go run main.go start

build:
	go build main.go -o ssm

info:
	@cat "$(CURDIR)/storage/database.json" && echo ""

fresh:
	rm storage/database.json
	rm storage/xray.json
	docker compose restart

update:
	git pull
	docker compose pull
	docker compose down
	rm ./storage/xray.json
	docker compose up -d
