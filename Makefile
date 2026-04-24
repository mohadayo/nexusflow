.PHONY: test test-python test-go test-ts lint up down build clean

test: test-python test-go test-ts
	@echo "All tests passed!"

test-python:
	@echo "==> Running Python tests..."
	cd gateway && pip install -q -r requirements.txt && pytest -v

test-go:
	@echo "==> Running Go tests..."
	cd engine && go test -v ./...

test-ts:
	@echo "==> Running TypeScript tests..."
	cd dashboard && npm install --silent && npm test

lint:
	@echo "==> Linting Python..."
	cd gateway && flake8 app.py test_app.py --max-line-length=120
	@echo "==> Linting TypeScript..."
	cd dashboard && npx eslint src/

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v --rmi local
	rm -rf dashboard/node_modules dashboard/dist
	rm -rf gateway/__pycache__
	rm -f engine/engine
