## Netchecker
Netchecker is a simple program to see if you can connect to a given
hostname:port combination. The code attempts to support UDP tests.
Testing UDP connectivity is iffy at best and the test will only pass
if the UPD service responds with anything.

####Usage:

```netchecker -f /path/to/config.yaml```

Here is an example of the configuration.

```
$> cat config.yaml
tcp:
  - "192.168.33.10:8300"
  - "192.168.33.10:8301"
  - "192.168.33.10:8302"
  - "192.168.33.10:8400"
  - "192.168.33.10:8500"
  - "192.168.33.10:8600"
udp:
  - "192.168.33.10:8301"
  - "192.168.33.10:8302"
  - "192.168.33.10:8600"
timeout_seconds: 2
```
