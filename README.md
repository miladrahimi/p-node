# P-Node

## Documentation

### Installation

```shell
# Install the requirements
apt-get -y update
apt-get -y install make wget curl vim git openssl cron
if command -v ufw &> /dev/null; then sudo ufw disable; fi
```

```shell
# Install Docker
wget -O install-docker.sh https://get.docker.com
chmod +x install-docker.sh && ./install-docker.sh && rm install-docker.sh
```

```shell
# Install BBR
sudo sh -c 'echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf'
sudo sh -c 'echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf'
sudo sysctl -p
```

```shell
# Install P-Node
for ((i=1;;i++)); do [ ! -d "p-node-${i}" ] && break; done
git clone https://github.com/miladrahimi/p-node.git "p-node-${i}"
cd "p-node-${i}"
make setup
docker compose up -d
```

```shell
# Show information required for P-Manager
make info
```

### Update

Automatic updates are set up through cron jobs by default.
For earlier updates, run the command below:

``` shell
make update
```

### Requirements

* Operating systems: Debian or Ubuntu
* Architecture: AMD64
* RAM: 1 GB or more
* CPU: 1 Core or more

## Links

* https://github.com/miladrahimi/p-manager

## License

This project is governed by the terms of the [LICENSE](LICENSE.md).
