.PHONY: install run build settings fresh

install:
	./third_party/install-xray.sh

run: install
	go run main.go start

build: install
	go build main.go -o ssm

settings:
	@json_file="$(CURDIR)/storage/database.json"; \
	http_port=$$(jq -r '.settings.http_port' "$$json_file"); \
	http_token=$$(jq -r '.settings.http_token' "$$json_file"); \
	echo "HTTP Port: $${http_port}"; \
	echo "HTTP Token: $${http_token}";

fresh:
	rm storage/database.json
	rm storage/xray.json
	docker compose restart
