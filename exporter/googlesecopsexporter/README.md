# Google SecOps Exporter

This exporter facilitates the sending of logs to Google SecOps, which is a security analytics platform provided by Google. It is designed to integrate with OpenTelemetry collectors to export telemetry data such as logs to a Google SecOps account.

## Supported APIs

The exporter supports two Google SecOps ingestion APIs:

- **Legacy Ingestion API (v2)**: Uses gRPC transport with malachite endpoints. This is the default protocol.
- **DataPlane API (v1alpha)**: Uses HTTPS transport with regional chronicle.googleapis.com endpoints.

The protocol is configured using the `protocol` field (see Configuration section below).

## Supported Pipelines

- Logs

## Prerequisites

Before using this exporter, you need:

1. **Google Cloud Service Account**: A service account with appropriate permissions to access the Google SecOps API
2. **Service Account Credentials**: JSON credentials file for the service account
3. **Google SecOps Access**: Your service account must have access to the Google SecOps API endpoints
4. **Customer ID**: Your Google SecOps customer ID (required for most configurations)
5. **API Selection**: Choose between the Legacy Ingestion API (gRPC) or the newer DataPlane API (HTTPS)

For more information on setting up credentials and API access, see the [Google SecOps API documentation](https://cloud.google.com/chronicle/docs/reference/ingestion-api).

## How It Works

1. The exporter uses the configured credentials to authenticate with the Google Cloud services.
2. It marshals logs into the format expected by Google SecOps.
3. It sends the logs to the appropriate Google SecOps endpoint.

## Configuration

The exporter can be configured using the following fields:

| Field                           | Type              | Default                                | Required | Description                                                                                                               |
| ------------------------------- | ----------------- | -------------------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------- |
| `protocol`                      | string            | `grpc`                                 | `false`  | The protocol/API to use: `grpc` (Legacy Ingestion API v2) or `https` (DataPlane API v1alpha).                            |
| `endpoint`                      | string            | `malachiteingestion-pa.googleapis.com` | `false`  | The endpoint for sending to Google SecOps (Legacy Ingestion API). See regional endpoints in the documentation.           |
| `location`                      | string            |                                        | `false`  | The GCP location/region (required when `protocol` is `https`).                                                           |
| `project`                       | string            |                                        | `false`  | The GCP project ID (required when `protocol` is `https`).                                                                |
| `creds_file_path`               | string            |                                        | `true`   | The file path to the Google credentials JSON file.                                                                        |
| `creds`                         | string            |                                        | `true`   | The Google credentials JSON.                                                                                              |
| `log_type`                      | string            |                                        | `false`  | The type of log that will be sent.                                                                                        |
| `raw_log_field`                 | string            |                                        | `false`  | The field name for raw logs.                                                                                              |
| `customer_id`                   | string            |                                        | `false`  | The customer ID used for sending logs.                                                                                    |
| `override_log_type`             | bool              | `true`                                 | `false`  | Whether or not to override the `log_type` in the config with `attributes["log_type"]`                                     |
| `namespace`                     | string            |                                        | `false`  | User-configured environment namespace to identify the data domain the logs originated from.                               |
| `compression`                   | string            | `none`                                 | `false`  | The compression type to use when sending logs. valid values are `none` and `gzip`                                         |
| `ingestion_labels`              | map[string]string |                                        | `false`  | Key-value pairs of labels to be applied to the logs when sent to Google SecOps.                                           |
| `collect_agent_metrics`         | bool              | `true`                                 | `false`  | Enables collecting metrics about the agent's process and log ingestion metrics                                            |
| `batch_request_size_limit_grpc` | int               | `4000000`                              | `false`  | The maximum size, in bytes, allowed for a gRPC batch creation request (Legacy Ingestion API).                             |
| `batch_request_size_limit_http` | int               | `4000000`                              | `false`  | The maximum size, in bytes, allowed for a HTTP batch creation request (DataPlane API).                                    |

### Log Type

If the `attributes["log_type"]` field is present in the log, and maps to a known Google SecOps `log_type` the exporter will use the value of that field as the log type. If the `attributes["log_type"]` field is not present, the exporter will use the value of the `log_type` configuration field as the log type.

currently supported log types are:

- windows_event.security
- windows_event.custom
- windows_event.application
- windows_event.system
- sql_server

If the `attributes["secops_log_type"]` or `attributes["chronicle_log_type"]` field is present in the log, we will use its value in the payload instead of the automatic detection or the `log_type` in the config. If both are present, `secops_log_type` takes priority.

### Namespace and Ingestion Labels

If the `attributes["secops_namespace"]` or `attributes["chronicle_namespace"]` field is present in the log, we will use its value in the payload instead of the `namespace` in the config. If both are present, `secops_namespace` takes priority.

If there are nested fields in `attributes["secops_ingestion_label"]` or `attributes["chronicle_ingestion_label"]`, we will use the values in the payload, merged with the `ingestion_labels` in the config. If both prefixes are present, they are merged with `secops_ingestion_label` values taking priority over `chronicle_ingestion_label` values for the same key.

## Credentials

This exporter requires a Google Cloud service account with access to the Google SecOps API. The service account must have access to the endpoint specfied in the config.
Besides the default endpoint, there are also regional endpoints that can be used [here](https://cloud.google.com/chronicle/docs/reference/ingestion-api#regional_endpoints).

For additional information on accessing Google SecOps, see the [Google SecOps documentation](https://cloud.google.com/chronicle/docs/reference/ingestion-api#getting_api_authentication_credentials).

## Log Batch Creation Request Limits

`batch_request_size_limit_grpc` and `batch_request_size_limit_http` are used for ensuring log batch creation requests don't exceed Google SecOps backend limits. These limits apply to the Legacy Ingestion API (gRPC) and DataPlane API (HTTPS) respectively. If a request exceeds the configured size limit, the request will be split into multiple requests that adhere to this limit, with each request containing a subset of the logs contained in the original request. Any single logs that result in the request exceeding the size limit will be dropped.

## Example Configuration

### Legacy Ingestion API (gRPC) - Basic Configuration

This is the default protocol using the Legacy Ingestion API (v2) with gRPC transport:

```yaml
googlesecops:
  protocol: grpc  # Optional - this is the default
  creds_file_path: "/path/to/google/creds.json"
  log_type: "ABSOLUTE"
  customer_id: "customer-123"
```

### Legacy Ingestion API (gRPC) - With Regional Endpoint

```yaml
googlesecops:
  protocol: grpc
  endpoint: malachiteingestion-eu.googleapis.com  # European endpoint
  creds_file_path: "/path/to/google/creds.json"
  log_type: "ONEPASSWORD"
  customer_id: "customer-123"
```

### DataPlane API (HTTPS) - Basic Configuration

Using the newer DataPlane API (v1alpha) with HTTPS transport:

```yaml
googlesecops:
  protocol: https
  location: us  # GCP region
  project: my-gcp-project-id
  creds_file_path: "/path/to/google/creds.json"
  log_type: "ABSOLUTE"
  customer_id: "customer-123"
```

### Configuration with Ingestion Labels

```yaml
googlesecops:
  protocol: grpc
  creds_file_path: "/path/to/google/creds.json"
  log_type: ""
  customer_id: "customer-123"
  ingestion_labels:
    env: dev
    zone: USA
```

## Additional Resources

- [Google SecOps Ingestion API Documentation](https://cloud.google.com/chronicle/docs/reference/ingestion-api)
- [Google SecOps API Authentication](https://cloud.google.com/chronicle/docs/reference/ingestion-api#getting_api_authentication_credentials)
- [Regional Endpoints](https://cloud.google.com/chronicle/docs/reference/ingestion-api#regional_endpoints)
- [Google SecOps Overview](https://cloud.google.com/products/security-operations)
