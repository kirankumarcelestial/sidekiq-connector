TAG?=1.0.2
.PHONY: build

build:
	docker build -t kirankumarcelestial/sidekiq-connector:$(TAG) .
push:
	docker push kirankumarcelestial/sidekiq-connector:$(TAG)
