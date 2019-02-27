export CLIENT=$(curl -s -XPOST localhost:8080/v1/create/client | jq .ID | sed -e 's/"//g')
export TOKEN=$(curl -s -XPOST localhost:8080/v1/create/token/$CLIENT | jq .ID | sed -e 's/"//g')
