# Httpd Receiver

This receiver can fetch stats from a Httpd instance using the 

> :construction: This receiver is currently in **BETA**.

## Details

This receiver supports Apache httpd version 2.4

## Configuration

### Httpd module

In order to receive server statistics, you must configure the servers `httpd.conf` file to [enable status support](https://httpd.apache.org/docs/2.4/mod/mod_status.html).


### Receiver Config

> :information_source: This receiver is in beta and configuration fields are subject to change.

The following settings are required:

- `endpoint` (default: `http://localhost:8080/server-status?auto`): The URL of the httpd status endpoint

The following settings are optional:

- `collection_interval` (default = `10s`): This receiver runs on an interval.
Each time it runs, it queries httpd, creates metrics, and sends them to the
next consumer. The `collection_interval` configuration option tells this
receiver the duration between runs. This value must be a string readable by
Golang's `ParseDuration` function (example: `1h30m`). Valid time units are
`ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

example:

```yaml
receivers:
  httpd:
    endpoint: "http://localhost:8080/server-status?auto"
    collection_interval: 10s
```
