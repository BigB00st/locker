# Locker

With locker, you can run linux containers.

## Prerequisites
Required:
* go-pie

Optional:
* apparmor
* iproute2
* iptables

## Installation
Locker is still in development. In addition, locker needs to make changes to your network interfaces, routing table, and firewall rules. Therefore, run at your own risk.

### From Source
```
git clone git@gitlab.com:amit-yuval/locker.git
make
sudo make install
```
### Arch Linux

```
git clone git@gitlab.com:amit-yuval/locker-aur.git
makepkg -s
sudo pacman -U locker-{VER}-{PKGREL}-x86_64.pkg.tar.xz
```

## Authors

* **Amit Botzer** - [BigB00st](https://github.com/BigB00st)
* **YuvalDahn** - [YuvalDahn](https://github.com/YuvalDahn)

## License
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)    
This project is licensed under the GPL License, version 3 - see the [LICENSE](LICENSE) file for details.