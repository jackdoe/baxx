dir=$(realpath $(dirname $0))
id=scylla-01
data=/var/lib/$id
mkdir -p $data/data $data/commitlog

docker run \
       --privileged \
       --name $id \
       --net=host \
       --restart on-failure \
       -v $data:/var/lib/scylla \
       -v $dir/scylla.yaml:/etc/scylla/scylla.yaml \
       -v /root/cert/db.crt:/etc/scylla/db.crt \
       -v /root/cert/db.key:/etc/scylla/db.key \
       -v /root/cert/cadb.pem:/etc/scylla/cadb.pem \
       -dit scylladb/scylla \
       --smp 4 \
       --memory 16G \
       --seeds "95.217.32.97,95.217.32.98" \
       --developer-mode 0 \
       --listen-address $(hostname -I | cut -f 1 -d ' ') \
       --broadcast-address $(hostname -I | cut -f 1 -d ' ')
