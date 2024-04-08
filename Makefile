.PHONY: all run server proxy system up down cagen fetch

APP_PATH = ./cmd/app/main.go
PROXY_PATH = ./cmd/proxy/main.go
CA_SCRIPT_PATH = ./scripts/ca.sh
CA_FILE = ./certs/ca.crt
CA_KEY = ./certs/ca.key
PARAMS_URL = https://raw.githubusercontent.com/PortSwigger/param-miner/master/resources/params

all: run

run: server proxy

server:
	go run $(APP_PATH)

proxy:
	go run $(PROXY_PATH) --ca_cert_file="$(CA_FILE)" --ca_key_file="$(CA_KEY)"

system:
	sudo systemctl start docker

up:
	sudo docker-compose up --remove-orphans

down:
	sudo docker-compose down --remove-orphans

cagen:
	sh $(CA_SCRIPT_PATH)

fetch:
	rm -rf resources/params
	wget $(PARAMS_URL) -P resources/
