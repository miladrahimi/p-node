.PHONY: setup run build info fresh update

setup:
	./third_party/install-xray-mac.sh

run:
	go run main.go start

build:
	go build main.go -o ssm

info:
	@json_file="$(CURDIR)/storage/database.json"; \
	http_port=$$(jq -r '.settings.http_port' "$$json_file"); \
	http_token=$$(jq -r '.settings.http_token' "$$json_file"); \
	echo "HTTP Port: $${http_port}"; \
	echo "HTTP Token: $${http_token}";

fresh:
	rm storage/database.json
	rm storage/xray.json
	docker compose restart

update:
	docker compose pull
	git pull
	docker compose down
	docker compose up -d
