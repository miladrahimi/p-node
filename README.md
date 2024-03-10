# Xray Node

> [!CAUTION]
> 
> **Update [Xray Manager](https://github.com/miladrahimi/xray-manager) to `v1.2.*` first!**
> 
> **Just run `make update && make update` in your [Xray Manager](https://github.com/miladrahimi/xray-manager) directory on your bridge server.**
>
> **Then check the version (run `make version`), it must be `v1.2.0` or newer.**
> 

## Documentation

### Installation

```shell
# Install the requirements
apt-get -y update
apt-get -y install make wget curl vim git openssl
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
# Install Xray Node
for ((i=1;;i++)); do [ ! -d "xray-node-${i}" ] && break; done
git clone https://github.com/miladrahimi/xray-node.git "xray-node-${i}"
cd "xray-node-${i}"
make setup
docker compose up -d
```

```shell
# Show information required for Xray Manager
make info
```

### Update

``` shell
# Execute this each time a new version is released
make update
```

### Requirements

 * Operating system: Linux (Debian or Ubuntu)
 * RAM: 1 GB or more
 * CPU: 1 Core or more

## Links

* https://github.com/miladrahimi/xray-manager

## License

Xray Node is released under the [MIT License](LICENSE).
