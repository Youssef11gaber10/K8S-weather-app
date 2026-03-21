# Weather App on Kubernetes

This repository is a learning project that ports a multi-service “weather” application from a **Docker Compose** style workflow to **Kubernetes**. It packages a Node.js UI, a Flask weather API, a Go authentication API, and MySQL into declarative manifests, with **NGINX Ingress** as the single HTTPS entry point.

![Kubernetes architecture overview](result-images/k8s-architecture-overview.png)

---

## What this project demonstrates

- **Multi-service app** split into Deployments and **ClusterIP** Services so the cluster handles discovery and load balancing across replicas.
- **Ingress** (`ingressClassName: nginx`) for **TLS** and external routing to the UI—no `proxy_pass` / `upstream` / SSL inside the UI container for those concerns.
- **StatefulSet + PersistentVolumeClaim** for MySQL data that must survive pod restarts.
- **Headless Service** for stable DNS to the MySQL pod (same idea as using a Compose service name as `DB_HOST`, expressed as Kubernetes DNS).
- **Jobs** for one-off initialization (schema/user grants) against the database.
- **Secrets** for passwords, JWT signing material, and API keys; **liveness/readiness** probes and **resource requests/limits** on workloads.

---

## Repository layout

```text
WeatherApp-K8S/
├── UI/                      # Node.js (Express) — static pages + server-side calls to auth & weather
├── weather/                 # Flask API — city weather (uses external API key from Secret)
├── authentication/          # Go API — users, JWT (talks to MySQL)
├── mysql-init/              # Reference SQL (optional; init also covered by Job + env in StatefulSet)
├── k8s/                     # Kubernetes manifests (see below)
└── result-images/           # Screenshots + architecture image for docs / portfolio
```

---

## Kubernetes manifests (`k8s/`)

| Path | Kind | Role |
|------|------|------|
| `k8s/UI/ingress.yml` | Ingress | Single host `weatherapp.local.com`, path `/` → `ui-service`; **TLS** via `tls-secret`. Uses `ingressClassName: nginx` (cluster must have NGINX Ingress Controller installed). |
| `k8s/UI/ui-deployment.yml` | Deployment | **2 replicas** of the UI image; env points `WEATHER_HOST` / `WEATHER_PORT` and `AUTH_HOST` / `AUTH_PORT` at in-cluster Services; JWT verify key from `mysql-secret`. Probes on `/health`. |
| `k8s/UI/ui-service.yml` | Service | **ClusterIP** — only reached from inside the cluster (and via Ingress from outside). |
| `k8s/weather/weather-deployment.yml` | Deployment | **2 replicas** of Flask; **RollingUpdate** strategy; API key from Secret `weather` (`rapid_api_key`). |
| `k8s/weather/weather-service.yml` | Service | ClusterIP → weather pods on port **5000**. |
| `k8s/auth/auth-deployment.yml` | Deployment | **2 replicas** of Go auth; DB and JWT env from `mysql-secret`; `DB_HOST` = headless service name. |
| `k8s/auth/auth-service.yml` | Service | ClusterIP → auth pods on port **8080**. |
| `k8s/auth/mysql/mysql-statefulset.yml` | StatefulSet | **1** MySQL pod; `volumeClaimTemplates` for **5Gi** PVC; env creates DB/user (alternative to only using a Job). |
| `k8s/auth/mysql/mysql-headless-service.yml` | Service | `clusterIP: None` — stable DNS per pod (`mysql-headless-service` resolves to MySQL pod IP(s)). |
| `k8s/auth/mysql/init-mysql-job.yml` | Job | One-shot client pod using `mysql` image to `CREATE DATABASE` / user / grants (same pattern as `docker compose` “run SQL once” against a service name). |

**TLS files:** `k8s/UI/tls/` holds example cert material for local use. In real clusters, create the Ingress TLS secret with `kubectl create secret tls tls-secret ...` (or cert-manager) instead of committing private keys.

---

## How traffic flows (this cluster vs “all paths in Ingress”)

In many tutorials, **Ingress** maps `/` to the frontend and `/api` to a backend. Here, **only the UI** is behind Ingress on `/`. The UI pod is still an **application server**: it calls **auth** and **weather** over **ClusterIP Services** using Kubernetes DNS (e.g. `weather-service`, `auth-service.default.svc.cluster.local`). So:

- **Ingress** = TLS + route to the UI Service.
- **Services** = stable names + load spread across pod replicas for auth, weather, and UI.
- **No** duplicate listing of backends like `backend1`, `backend2`, `backend3` in config—**scale `replicas` on the Deployment** and the Service keeps balancing.

---

## Docker Compose vs Kubernetes — how the same responsibilities move

| Concern | Typical Docker Compose approach | In this Kubernetes project |
|--------|--------------------------------|------------------------------|
| **Publish app on HTTPS** | One NGINX container: `listen 443`, certs, `proxy_pass` | **Ingress** terminates TLS and forwards to `ui-service` |
| **Load balance across copies** | Explicit `upstream { server app1; server app2; ... }` or multiple service names | **Service** with **selector** → all pods with that label; increase **replicas** without new hostnames |
| **Stable DB hostname** | Service name as `DB_HOST` | **Headless Service** + **StatefulSet** for MySQL (`mysql-headless-service`) |
| **Scale** | Add `backend2`, `backend3`, … to Compose | Change `replicas:` on Deployment |
| **Config / secrets** | `.env`, bind mounts | **Secrets** + `env.valueFrom.secretKeyRef` in Pod specs |
| **One-off DB setup** | Entrypoint scripts or manual `mysql` client | **Job** that runs SQL against the DB Service DNS name |

**Why a cloud load balancer can sit in front of Ingress (optional):** The Ingress controller runs **inside** the cluster. On cloud providers, a **LoadBalancer**-type Service (often installed with the ingress controller chart) gets a **public IP or hostname** from the cloud so Internet traffic can reach that controller. On bare metal / local clusters you might use **MetalLB**, **NodePort**, or **port-forward** instead. The Ingress resource still defines *rules* (host, path, TLS); the LB is *how* packets reach the controller.

---

## What I learned (short)

1. **Compose “many containers = many service names”** maps to **one Service + many Pods** in Kubernetes—scaling is a number, not a new block in the file.
2. **NGINX inside the app image** can shrink to **static + app logic**; **routing, TLS, and class-based ingress** move to **Ingress** and the **ingress controller**.
3. **Headless Services** match the mental model of “resolve this name to my database pod,” similar to Compose DNS for a DB service.
4. **StatefulSet + PVC** is the natural place for **data** that must outlive a single container.

---

## Application screenshots

Screenshots below show the running app: browser trust for self-signed TLS, signup, login, and weather lookup.

| Step | Preview |
|------|---------|
| HTTPS (certificate) | ![SSL](result-images/1-ssl.png) |
| Sign up | ![Signup](result-images/2-signup.png) |
| Login | ![Login](result-images/3-login.png) |
| Weather | ![Weather](result-images/4-weather.png) |

---

## Prerequisites (high level)

- A Kubernetes cluster with **NGINX Ingress Controller** installed and an **IngressClass** named `nginx` (or adjust `ingressClassName` in `ingress.yml`).
- **StorageClass** `standard` (or change `storageClassName` in the MySQL StatefulSet) for the PVC.
- **Secrets** before applying manifests, for example:
  - `mysql-secret`: keys such as `root-password`, `auth-password`, `secret-jwt` (as referenced in the Deployments).
  - `weather`: key `rapid_api_key` for the weather Deployment.
  - `tls-secret`: TLS cert/key for host `weatherapp.local.com` for the Ingress.
- Local DNS or `/etc/hosts` entry: `weatherapp.local.com` → Ingress external IP.

Images referenced in manifests are placeholders from the author’s registry; replace with your own builds from `UI/`, `weather/`, and `authentication/` Dockerfiles.

---
