# Exemplars demo

Docker based demo application for generating Prometheus exemplar data that generates traces in an Azure Application Insights workspace.

## Requirements

Docker Compose is required to run the containers in this example.

The following environment variables are also required:

- `APPLICATIONINSIGHTS_CONNECTION_STRING` - The connection string for the Application Insights workspace that traces will be written to.
- `TENANT_ID` - The tenant ID for an App Registration that will be used to authenticate to Azure.
- `CLIENT_ID` - The client ID for an App Registration that will be used to authenticate to Azure.
- `CLIENT_SECRET` - The client secret for an App Registration that will be used to authenticate to Azure.

Export the above environment variables and then run `./start.sh` from a terminal. The docker containers will be created and Grafana will be available on `localhost:3002`.

The Grafana instance will be preconfigured with Azure Monitor, Prometheus, Loki, and Tempo data sources.

These data sources are appropriately configured to connect to the relevant container instances where the generated data can be viewed.

To stop the demo application run `./stop.sh`.
