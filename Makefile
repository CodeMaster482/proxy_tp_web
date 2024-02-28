
system:
	sudo systemctl start docker

up:
	sudo docker-compose up --remove-orphans

down:
	sudo docker-compose down --remove-orphans

cagen:
	sh  ./scripts/ca.sh
