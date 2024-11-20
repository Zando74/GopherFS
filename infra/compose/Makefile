start_cluster:
	docker-compose -f ./compose.dev.yml up --build -d control-plane-1
stop_cluster:
	docker-compose -f ./compose.dev.yml down

start_leader:
	docker-compose -f ./compose.dev.yml up --build -d control-plane-1
stop_leader:
	docker-compose -f ./compose.dev.yml down control-plane-1

start_follower_1:
	docker-compose -f ./compose.dev.yml up --build -d control-plane-2
stop_follower_1:
	docker-compose -f ./compose.dev.yml down control-plane-2

start_follower_2:
	docker-compose -f ./compose.dev.yml up --build -d control-plane-3
stop_follower_2:
	docker-compose -f ./compose.dev.yml down control-plane-3