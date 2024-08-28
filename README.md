# insights-client

A new generation client for Red Hat Insights.

---

Red Hat Insights is a SaaS platform for RHEL management. insights-client manages collection of data on a host and uploads it to Insights for analysis.

**References:**

- [Red Hat Insights](https://consoledot.redhat.com/insights): Red Hat cloud services
- [insights-core](https://github.com/RedHatInsights/insights-core): Set of collectors that gather host data
- [insights-client](https://github.com/RedHatInsights/insights-client): A lightweight system wrapper around Core


## Development

### Building and running

```shell
make build
./bin/insights-client
```

### Testing

```shell
make check
```

### Environment variables

The use of environment variables to change the behavior should be limited and always constraint to development or testing purposes.
All production configuration should be made via the configuration file.

- `HTTP_DEBUG`: Include more detailed information about HTTP traffic (e.g. raw responses). Commonly used with `--debug`.

## Contributing

This project is developed under the [MIT license](LICENSE).

See [CONTRIBUTING.md](CONTRIBUTING.md) to learn more about the contribution process, Conventional Commits and Developer Certificate of Origin.
