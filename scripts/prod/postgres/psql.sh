source /root/.pass
docker run -e PGPASSWORD="$POSTGRES_PASSWORD" --rm -it --net=host postgres:11-alpine psql -h 127.0.0.1 -U root baxx