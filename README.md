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

## Payments SDK
The sdk is a Go wrapper around the listed payments apis, which provides a simple interface to interact with them.

It has the following unique features:
- **Request Hooks**: powerful request hook design which unlocks the ability to extend the sdk with custom hooks that 
that meet unique business cases. Build custom hooks to intercept and modify requests before they are sent as well as hooks
intercept and modify responses.
- **Request Retrier**: ready to use retrier with exponential backoff and jitter.

### Installation
Use go get.
```bash
go get github.com/SirWaithaka/payments
```

Then import the payments sdk package into your code
```go
import "github.com/SirWaithaka/payments"
```

### Usage and Documentation
Please see examples for usage.
- [Simple Request](https://github.com/SirWaithaka/payments/blob/main/examples/simple/main.go)
- [Daraja C2B Request](https://github.com/SirWaithaka/payments/blob/main/examples/daraja/main.go)


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
This project has been inspired by problems and challenges I have faced while building a payments apis. Below I describe some
of the challenges I faced, most of them around enabling M-Pesa payments.

### 1. Process and Make M-Pesa payments
In Kenya M-Pesa has a [market share of 90%](https://techweez.com/2025/07/01/kenya-goes-cashless-mobile-money-subscriptions-soar-to-45-million/).
This means that for many businesses, a majority of their customer base will be using M-Pesa, and hence, need to support
M-Pesa payments. At the same time, the same business will also want to make their own payments as well as other accounting
needs. These businesses, as well as businesses building solutions to enable commerce, will need to integrate with M-Pesa.

### 2. Consuming different M-Pesa APIs for different products
This is attributed to compliance. Due to M-Pesa dominating the Kenyan market, a number of companies provide payment solutions
through M-Pesa, hence, some finance companies have partnerships with such companies as well as in-built legacy apis 
integrating with M-Pesa. This means, payment flows for products such as M-Pesa C2B Stk, B2C, B2B would be calling different APIs internally.

One example, is a requirement to enable support B2B transactions for a bank via the Daraja API, while keeping C2B and B2C
transactions going to a legacy API, and provide a unified interface for all of them. This poses a couple of challenges:
- How to handle different M-Pesa shortcodes for the different payment types
- How to handle transaction status calls to the different APIs and mapping the responses.

### 3. Redundant payment rails
M-Pesa has a core service called G2 which uses the legacy XML format, and Daraja is a REST wrapper around it. Quikk is
another solution that is a REST wrapper around G2. From time to time, albeit not often, the Daraja API faces intermittent
failures that affect transactions. Quikk can be used as a fallback for Daraja in such scenarios.

*It is important to note that Daraja is Free to use while Quikk is not.*

### 4. Replay and Fan-Out Webhooks
From time to time, clients may suffer downtime or network issues and won't be able to receive or process all payment
notifications on time. Equally, you could be adding a new client and require it to process payment notifications from a
particular period. Payment notifications received via webhooks from payment providers can be replayed or duplicated and 
sent to multiple http clients in a fan-out fashion.

