.PHONY: setup run build info fresh update

setup:
	./third_party/install-xray-mac.sh

run:
	go run main.go start

build:
	go build main.go -o ssm

info:
	@printf "IP: "
	@curl ifconfig.io
	@printf "DB: "
	@cat "$(CURDIR)/storage/database.json" && echo ""

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
