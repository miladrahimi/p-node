# Xray Node

## Documentation

### Installation

```shell
# Install the requirements
apt-get -y update && apt-get -y upgrade
apt-get -y install make wget curl vim git openssl

# Install Docker
wget -O install-docker.sh https://get.docker.com
chmod +x install-docker.sh && ./install-docker.sh

# Install BBR
sudo sh -c 'echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf'
sudo sh -c 'echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf'
sudo sysctl -p
```

```shell
# Install Xray Node
git clone https://github.com/miladrahimi/xray-node.git
cd xray-node
docker compose up -d
```

```shell
# Show Information required for Xray Manager
make info
```

### Update

``` shell
make update
# Execute this each time a new version is released.
```

## Links

* https://github.com/miladrahimi/xray-manager
