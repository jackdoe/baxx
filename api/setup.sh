#rm -f /tmp/gorm.sqlite3
rm -f /tmp/baxx/*
export CLIENT=$(curl -s -XPOST localhost:8080/v1/create/client | jq .ID | sed -e 's/"//g')
export TOKEN=$(curl -s -XPOST localhost:8080/v1/create/token/$CLIENT | jq .ID | sed -e 's/"//g')
curl -d '{"Type":"email","Value":"jack@prymr.nl"}' http://localhost:8080/v1/create/destination/$CLIENT
curl -d '{"Type":"delta%","Value":-1, "Destinations":[{"Type":"email","Value":"jack@prymr.nl"}], "Match": ".*"}' http://localhost:8080/v1/create/notification/$CLIENT/$TOKEN
cat main.go| curl -XPOST --data-binary @- localhost:8080/v1/upload/$CLIENT/$TOKEN/todzzsz | jq
curl localhost:8080/v1/download/$CLIENT/$TOKEN/todzzsz
