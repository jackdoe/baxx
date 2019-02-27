export CLIENT=$(curl -s -XPOST localhost:8080/v1/create/client | jq .ID | sed -e 's/"//g')
export TOKEN=$(curl -s -XPOST localhost:8080/v1/create/token/$CLIENT | jq .ID | sed -e 's/"//g')
curl -d '{"Type":"email","Value":"jack@prymr.nl"}' http://localhost:8080/v1/create/destination/$CLIENT
 curl -d '{"Type":"delta%","Value":-1, "Destinations":[{"Type":"email","Value":"jack@prymr.nl"}]}' http://localhost:8080/v1/create/notification/$CLIENT/$TOKEN
