#rm -f /tmp/gorm.sqlite3
rm -f /tmp/baxx/*
export CLIENT=$(curl -s -XPOST -d "{\"email\":\"jack@prymr.nl\", \"password\":\"asd\"}" localhost:8080/v1/register | jq .secret | sed -e 's/"//g')
export TOKEN=$(curl -ujack@prymr.nl:asd -s -d '{"WriteOnly":false, "NumberOfArchives":0}' -XPOST -d '{}' localhost:8080/protected/v1/create/token | jq .token | sed -e 's/"//g')
cat main.go | pv -L 10K -p | curl -XPOST --data-binary @- localhost:8080/v1/io/$CLIENT/$TOKEN/todzzsz | jq
curl localhost:8080/v1/io/$CLIENT/$TOKEN/todzzsz
echo 123 | curl -XPOST --data-binary @- localhost:8080/v1/io/$CLIENT/$TOKEN/todzzsz | jq
echo 125 | curl --data-binary @- localhost:8080/v1/io/$CLIENT/$TOKEN/todzzsz | jq
echo 125 | curl -XPOST --data-binary @- localhost:8080/v1/io/$CLIENT/$TOKEN/todzzsz | jq
echo 126 | curl -XPOST --data-binary @- localhost:8080/v1/io/$TOKEN/todzzsz | jq
curl localhost:8080/v1/io/$CLIENT/$TOKEN/todzzsz


export TOKEN=$(curl -ujack@prymr.nl:asd -s -d '{"WriteOnly":true, "NumberOfArchives":0}' -XPOST localhost:8080/protected/v1/create/token | jq .token | sed -e 's/"//g')
echo 123 | curl --data-binary @- localhost:8080/v1/io/$CLIENT/$TOKEN/todzzsz | jq
curl localhost:8080/v1/io/$CLIENT/$TOKEN/todzzsz | jq
curl -ujack@prymr.nl:asd localhost:8080/protected/v1/io/$CLIENT/$TOKEN/todzzsz
