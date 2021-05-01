docker build -t postgresdb .
docker run -d -P --name postgrescontainer postgresdb
