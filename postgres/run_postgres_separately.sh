docker build -t postgresdb .
docker run -d -P -p 5432:5432 --name postgrescontainer postgresdb
