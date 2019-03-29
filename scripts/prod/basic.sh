# super basic setup docker and ipfw

apt-get update
apt-get upgrade
apt-get install apt-transport-https ca-certificates curl gnupg-agent ntp git software-properties-common ufw
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
add-apt-repository  "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs)  stable"
apt-get update
apt-get install docker-ce

ufw default deny outgoing comment 'deny all outgoing traffic'
ufw default deny incoming comment 'deny all incoming traffic'

ufw limit in ssh comment 'allow SSH connections in'

ufw allow in http comment 'allow HTTP traffic in'
ufw allow in https comment 'allow HTTPS traffic in'

ufw allow out 53 comment 'allow DNS calls out'
ufw allow out 123 comment 'allow NTP out'

ufw allow out http comment 'allow HTTP traffic out'
ufw allow out https comment 'allow HTTPS traffic out'


ufw allow from 95.217.32.97/32 to 95.217.32.98/32 port 7000 comment 'allow scylla IPC from 95.217.32.97->98'
ufw allow from 95.217.32.98/32 to 95.217.32.97/32 port 7000 comment 'allow scylla IPC from 95.217.32.98->97'

ufw enable

# otherwise docker hijacks it and allows traffic
echo 'DOCKER_OPTS="--iptables=false"' >> /etc/default/docker