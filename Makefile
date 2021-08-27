
dev-up:
	@docker-compose up -d

dev-down:
	@docker-compose down

rebuild:
	@docker-compose up -d --no-deps --build canal-kafka-connector  kafka-es-connector producer
