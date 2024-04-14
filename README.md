# P-Node

## Documentation

### Installation

1. Install the requirements

```shell
apt-get -y update
apt-get -y install make wget curl vim git openssl cron
if command -v ufw &> /dev/null; then sudo ufw disable; fi
```

2. Install Docker

```shell
wget -O install-docker.sh https://get.docker.com
chmod +x install-docker.sh && ./install-docker.sh && rm install-docker.sh
```

3. Install BBR

```shell
sudo sh -c 'echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf'
sudo sh -c 'echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf'
sudo sysctl -p
```

4. Install P-Node

```shell
for ((i=1;;i++)); do [ ! -d "p-node-${i}" ] && break; done
git clone https://github.com/miladrahimi/p-node.git "p-node-${i}"
cd "p-node-${i}"
make setup
docker compose up -d
```

5. Show information required for P-Manager

```shell
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
* Architecture: `amd64`
* RAM: 1 GB or more
* CPU: 1 Core or more

## Links

* [P-Manager](https://github.com/miladrahimi/p-manager)

## License

This project is governed by the terms of the [LICENSE](LICENSE.md).
