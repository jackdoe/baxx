#!/bin/bash
dir=$(realpath $(dirname $0))
source /root/.pass

docker stop postgres
docker rm postgres

docker run \
       --net=host \
       --name postgres-slave \
       -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
       -e POSTGRES_USER=$POSTGRES_USER \
       -e POSTGRES_DB=$POSTGRES_DB \
       -v $dir/postgres.conf:/etc/postgresql/postgresql.conf \
       -v $dir/hba.conf:/var/lib/postgresql/data/pg_hba.conf \
       -v /root/cert/pg/server.crt:/ca/server.crt:ro \
       -v /root/cert/pg/server.crt:/ca/server.crt:ro \
       -v /root/cert/pg/server.key:/ca/server.key:ro \
       -v /var/lib/pgdata:/var/lib/postgresql/data \
       postgres:11.2 \
       -c ssl=on \
       -c ssl_cert_file=/ca/server.crt \
       -c ssl_key_file=/ca/server.key \
       -c 'config_file=/etc/postgresql/postgresql.conf'
