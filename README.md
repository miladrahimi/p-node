# P-Node

P-Node is a lightweight Xray proxy node designed to work seamlessly with P-Manager.

## Documentation

### Installation

Run the one-line installer as root on a fresh Debian/Ubuntu server:

```shell
curl -fsSL https://raw.githubusercontent.com/miladrahimi/p-node/master/scripts/install.sh | sudo bash
```

Or with `wget`:

```shell
wget -qO- https://raw.githubusercontent.com/miladrahimi/p-node/master/scripts/install.sh | sudo bash
```

It installs the requirements, sets up the service, and prints the info required
for P-Manager. `make info` output has three JSON variants — `Full` (Xray push +
SSH), `Xray-only`, and `SSH-only`; paste the one matching how P-Manager should
use this node into the P-Manager "Add Node" input.

<details>
<summary>Manual installation (advanced)</summary>

1. Install the requirements

```shell
apt-get -y update
apt-get -y install make wget jq curl vim git openssl cron
```

2. Install BBR (Optional)

```shell
echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf
echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf
sysctl -p
```

3. Install P-Node

```shell
for ((i=1;;i++)); do [ ! -d "p-node-${i}" ] && break; done
git clone https://github.com/miladrahimi/p-node.git "p-node-${i}"
cd "p-node-${i}"
make setup
```

4. Display information required for P-Manager

```shell
make info
```

</details>

### Update

Automatic updates are set up through cron jobs by default. For earlier updates, run the command below:

``` shell
make update
```

### Uninstall

The command below uninstalls the P-Node service, update cron job, and application directory.

```shell
sudo make uninstall
```

### Status and Logs

The application service is named after its directory, with `p-node` as the default in `systemd`.
It allows running multiple instances on a single server by placing the application in different directories.
Default naming strategy for multiple instances is using `p-node-2` and `p-node-3`, etc.

To check the status of the application, execute the following command:

```shell
systemctl status p-node-1
```

To view the application's standard outputs, execute the command below:

```shell
journalctl -f -u p-node-1
```

The application logs will be stored in the following directory:

```shell
./storage/logs
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
