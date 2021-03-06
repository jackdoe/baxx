# super basic setup docker and ipfw

apt-get update
apt-get upgrade
apt-get install apt-transport-https ca-certificates curl gnupg-agent ntp git software-properties-common ufw
snap install go --classic
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
add-apt-repository  "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs)  stable"
apt-get update -y
apt-get install -y docker-ce




ufw default deny outgoing comment 'deny all outgoing traffic'
ufw default deny incoming comment 'deny all incoming traffic'

ufw limit in ssh comment 'allow SSH connections in'

ufw allow in http comment 'allow HTTP traffic in'
ufw allow in https comment 'allow HTTPS traffic in'

ufw allow out 53 comment 'allow DNS calls out'
ufw allow out 123 comment 'allow NTP out'

ufw allow out 465 comment 'allow mail out' # XXX: allow only sendcrid

ufw allow out http comment 'allow HTTP traffic out'
ufw allow out https comment 'allow HTTPS traffic out'

# XXX: make it stricter, port by port
ufw allow in from 95.217.32.98/32
ufw allow in from 95.217.32.97/32
ufw allow out to 95.217.32.98/32
ufw allow out to 95.217.32.97/32

ufw enable

# fuck
# disable iptables
# limit the logs
# make logging non blocking

cat > /etc/docker/daemon.json <<EOL
{
  "iptables":false,
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3",
    "mode": "non-blocking",
    "labels": "production_status",
    "env": "os,customer"
  }
}
EOL

