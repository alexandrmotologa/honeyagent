# Deployment Guide

honeyagent compiles to a standalone binary with zero external runtime dependencies. It embeds its React 19 management interface into the binary and runs with less than 20MB of resident memory.

## Binary execution

Download or build the binary:

```bash
make build
./bin/honeyagent serve --config ./examples/honeyagent.yaml
```

The server opens two network ports:
- Port `7878`: Management REST API and embedded React 19 Deception Studio.
- Port `8088`: Airlock reverse proxy gateway forwarding to the upstream LLM service.

## Configuration options

Configure honeyagent with `honeyagent.yaml` or environment variables:

| Setting | Environment Variable | Default | Purpose |
|---|---|---|---|
| `server.port` | `HONEYAGENT_PORT` | `7878` | Management studio and API port |
| `server.proxy_port` | `HONEYAGENT_PROXY_PORT` | `8088` | Airlock reverse proxy port |
| `server.upstream_url` | `HONEYAGENT_UPSTREAM_URL` | `http://localhost:8000` | Target RAG or LLM server |
| `canary.master_key` | `HONEYAGENT_MASTER_KEY` | *(built-in default)* | Secret key used to sign HMAC tokens |
| `alerts.slack.webhook_url` | `HONEYAGENT_SLACK_WEBHOOK` | *(empty)* | Slack incoming webhook destination |
| `alerts.pagerduty.routing_key` | `HONEYAGENT_PAGERDUTY_KEY` | *(empty)* | PagerDuty service routing key |

## Kubernetes sidecar deployment

In a Kubernetes pod, run honeyagent as a sidecar container next to the primary RAG application container:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: enterprise-rag-service
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: rag-backend
          image: internal-registry/enterprise-rag:v2.4
          ports:
            - containerPort: 8000

        - name: honeyagent-airlock
          image: internal-registry/honeyagent:v1.0.0
          args: ["serve", "--upstream", "http://127.0.0.1:8000"]
          ports:
            - containerPort: 8088
            - containerPort: 7878
          resources:
            limits:
              memory: "64Mi"
              cpu: "250m"
            requests:
              memory: "20Mi"
              cpu: "50m"
```

Traffic routes through `honeyagent-airlock:8088`, inspecting incoming queries and outgoing context chunks before they leave the pod boundary.
