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
wget -N --no-check-certificate https://github.com/teddysun/across/raw/master/bbr.sh
chmod +x bbr.sh && bash bbr.sh
```

```shell
# Install Xray Node
git clone https://github.com/miladrahimi/xray-node.git xray-node
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
