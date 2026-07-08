SHELL=/bin/bash

.PHONY: start stop

# -------- Docker commands ----------- #
start:
	docker compose up -d --build app

stop:
	docker compose down -v