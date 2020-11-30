# go-mackerel-plugin-mysql-lite

Yet another mackerel plugin for MySQL.
Fetch replication and threads metrics only.

## Usage

```
Usage:
  mackerel-plugin-mysql-lite [OPTIONS]

Application Options:
  -H, --host=     Hostname (localhost)
  -p, --port=     Port (3306)
  -u, --user=     Username (root)
  -P, --password= Password

Help Options:
  -h, --help      Show this help message
```

Example

```
$ ./mackerel-plugin-mysql-lite
mysql-lite.replication-behind-master.second     0       1458026282
mysql-lite.replication-threads.io       0       1458026282
mysql-lite.replication-threads.sql      0       1458026282
mysql-lite.threads.running      1       1458026282
mysql-lite.threads.connected    1       1458026282
mysql-lite.threads.cached       0       1458026282
mysql-lite.threads.max-connections      151     1458026282
mysql-lite.threads.cache-size   0       1458026282
mysql-lite.connections.utilization      0.662252        1458026282
```

## Install

Please download release page or `mkr plugin install kazeburo/go-mackerel-plugin-mysql-lite`.

