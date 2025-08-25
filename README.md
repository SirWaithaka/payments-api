# Payments Api

This is both a payments api and a Go SDK for common Kenyan payment apis, which include payment processors, banks,
digital wallets and other fintechs.

## 🚧 Project Under Development 🚧

**DISCLAIMER:** This project is currently under heavy development. Features may be incomplete, unstable, or subject to 
significant changes without notice. Use at your own risk.

_Last updated: Aug 16, 2025_

## Roadmap
- [x] Daraja
  - [x] C2B Stk
  - [x] B2C
  - [x] B2B
  - [x] Transaction Status
  - [x] Account Balance
  - [ ] Reversal
  - [ ] Org Name check
- [x] Quikk
  - [x] C2B Stk
  - [x] B2C
  - [x] B2B
  - [x] Transaction Status
  - [x] Account Balance
  - [ ] Refund
- [ ] Tanda
- [ ] JamboPay
- [ ] Airtel Money
- [ ] Pesalink

## Getting Started with the API

The API can be deployed as a container or using Makefile as a binary. It has a dependency on Kafka and Postgres. 

### Setting up Kafka and Postgres
Use `docker-compose.infra.yml` file to configure and deploy kafka and postgres. It will start a zookeeper instance and expose
port 2181, a kafka instance on port 9092 and a postgres instance on port 5432. It will also create a docker network called
`apps` which is used by the payments api to connect to kafka and postgres.

```bash
# deploy kafka and postgres
docker compose -f docker-compose.infra.yml up -d
```

### Docker Compose
The compose file has 3 services:
- `migrations`: The database migrations.
- `seeds`: Database seeds. Run this only when testing the api.
- `payments-api`: The payments api.

The compose file requires a `.env` file with the necessary environment variables to connect to the database
instance. See below example:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DATABASE=payments
POSTGRES_SCHEMA=public
```

Build the necessary images and run the application

```bash
# build the images
docker compose build 
```
Run the api
```bash
# migrations will be run first, then seeds will be run and finally the api will be started
docker compose up -d payments-api
```

API will be exposed on port 6001. Check its reachable
```bash
curl -k localhost:6001/health
# response should be
OK
```

## Inspiration