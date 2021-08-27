
dev-up:
	@docker-compose up -d

rebuild:
	@docker-compose up -d --no-deps --build canal-kafka-connector  kafka-es-connector producer
