# Go + RabbitMQ Fan-out Demo

The Shop application with event publishing: every successful buy publishes an event to the RabbitMQ fanout exchange `product_events`, and two independent services each receive their own copy.

## Stack

- Go 1.27 with `github.com/rabbitmq/amqp091-go`
- RabbitMQ 4.4 (with management UI)
- PostgreSQL 18
- Vue.js 3 (frontend), Caddy
- Docker & Docker Compose

## What This Demo Shows

```
                                                   ┌→ email_queue       → email-service
Browser → Caddy → backend → product_events (fanout)┤
                     ↓                             └→ fulfillment_queue → fulfilment-service
                 PostgreSQL
```

- The backend publishes to the exchange, not to a queue, and does not know who consumes the event
- The exchange copies each message to all bound queues
- Each service declares and binds its own durable queue, so a new consumer needs no change in the backend
- Messages are published as persistent and acknowledged manually after processing

Only `POST /api/products/{id}/buy` publishes an event. The message body is the updated product as JSON. Creating a product publishes nothing.

The backend also declares `email_queue`, but publishes only to the exchange. A queue receives events only once its service has bound it, so events published before a service first started are not delivered to it.

## Running Locally

```bash
docker compose up --build
```

| Service | URL / Port | Notes |
|---------|------------|-------|
| Shop (Caddy) | `http://localhost:8081` | Serves the frontend, proxies `/api/*` to the backend |
| RabbitMQ management UI | `http://localhost:15672` | Login `guest` / `guest` |
| RabbitMQ AMQP | `localhost:5672` | |
| PostgreSQL | `localhost:5432` | |

## API Endpoints

- `GET /api/products` - List all products
- `POST /api/products` - Create new product
- `POST /api/products/{id}/buy` - Decrement stock and publish event

## See It Working

Follow the logs of both consumers:

```bash
docker compose logs -f email-service fulfilment-service
```

Add a product and click Buy in the frontend, or use curl:

```bash
curl -X POST http://localhost:8081/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Keyboard","description":"Mechanical","price":99.9,"stock":5}'
curl -X POST http://localhost:8081/api/products/1/buy
```

Both services log the same event: `Processing email: {...}` and `Processing fulfillment: {...}`.

In the management UI, the Exchanges tab shows `product_events` with both queues bound, and the Queues tab shows `email_queue` and `fulfillment_queue`.

**A stopped consumer catches up:** run `docker compose stop fulfilment-service` and buy something. Only email-service logs the event, and the message waits in `fulfillment_queue`. After `docker compose start fulfilment-service` it is processed.

**Competing consumers:** with `docker compose up --scale email-service=2`, both instances consume from `email_queue` and each message goes to only one of them.
