#rm -f /tmp/gorm.sqlite3
rm -f /tmp/baxx/*
export CLIENT=$(curl  curl -s -XPOST -d '{"email":"jack@prymr.nl", "password":"asd"}' localhost:8080/v1/register | jq .secret | sed -e 's/"//g')
export TOKEN=$(curl -ujack@prymr.nl:asd -s -d '{"WriteOnly":false, "NumberOfArchives":0}' -XPOST -d '{}' localhost:8080/v1/protected/create/token | jq .token | sed -e 's/"//g')
#curl -d '{"Type":"email","Value":"jack@prymr.nl"}' http://localhost:8080/v1/create/destination/$CLIENT
#curl -d '{"Type":"delta%","Value":-1, "Destinations":[{"Type":"email","Value":"jack@prymr.nl"}], "Match": ".*"}' http://localhost:8080/v1/create/notification/$CLIENT/$TOKEN
cat main.go| curl -XPOST --data-binary @- localhost:8080/v1/upload/$CLIENT/$TOKEN/todzzsz | jq
curl localhost:8080/v1/download/$CLIENT/$TOKEN/todzzsz
echo 123 | curl -XPOST --data-binary @- localhost:8080/v1/upload/$CLIENT/$TOKEN/todzzsz | jq
echo 125 | curl -XPOST --data-binary @- localhost:8080/v1/upload/$CLIENT/$TOKEN/todzzsz | jq
echo 125 | curl -XPOST --data-binary @- localhost:8080/v1/upload/$CLIENT/$TOKEN/todzzsz | jq
echo 126 | curl -XPOST --data-binary @- localhost:8080/v1/upload/$CLIENT/$TOKEN/todzzsz | jq
curl localhost:8080/v1/download/$CLIENT/$TOKEN/todzzsz


export TOKEN=$(curl -ujack@prymr.nl:asd -s -d '{"WriteOnly":true, "NumberOfArchives":0}' -XPOST -d '{}' localhost:8080/v1/protected/create/token | jq .token | sed -e 's/"//g')
echo 123 | curl -XPOST --data-binary @- localhost:8080/v1/upload/$CLIENT/$TOKEN/todzzsz | jq
curl localhost:8080/v1/download/$CLIENT/$TOKEN/todzzsz | jq
curl -ujack@prymr.nl:asd localhost:8080/v1/protected/download/$CLIENT/$TOKEN/todzzsz
