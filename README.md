# insights-client

## Running development version

```shell
$ make build
$ ./bin/insights-client --version
Client:     4.0.0a0+a9697fb
Collectors
- advisor:  3.3.23
```

If you want to run this with upstream Core, you have to export the path in `PYTHONPATH`.
Because the Core uses rhsm package to load information about API hostname, you may want to export path to subscription-manager as well.

```shell
$ make build
$ export PYTHONPATH=/path/to/insights-core:/path/to/subscription-manager/src
$ sudo ./bin/insights-client --collector SOMETHING
```

## Collector discovery

A collector file must be placed into `/etc/insights-client/collectors/` as a unique INI file:

```ini
[collector]
name = advisor
version = 3.3.23
exec = /usr/bin/python3 -m insights.anchor.advisor
content-type = application/vnd.redhat.advisor.collection+tgz
```

where `exec` is an executable file that performs the collection.

Your application is expected to save the archive as `ARCHIVE_PATH` which the Client specifies in environment.
