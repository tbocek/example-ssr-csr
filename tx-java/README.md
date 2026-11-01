# Spring Boot Transaction Management Demo

A minimal Spring Boot application demonstrating transaction management with `@Transactional`, showing the difference between atomic operations and non-atomic operations.

## Stack

- Java 27
- Spring Boot 4.1.2
- Spring Data JPA
- PostgreSQL 18
- Docker & Docker Compose

## Project Structure

```
src/main/java/ch/ost/i/dsl/tx/
├── DemoApplication.java                    # Main application entry point
├── Product.java                               # JPA entity for products
├── ProductRepository.java                     # JPA repository for products
├── ProductStatistics.java                     # JPA entity for statistics
├── ProductStatisticsRepository.java           # JPA repository for statistics
├── ProductService.java                        # Service layer with @Transactional
└── ProductController.java                     # MVC controller
```

## What This Demo Shows

**The Problem:** When multiple database operations must succeed together (atomicity), exceptions can leave data in an inconsistent state.

**The Solution:** `@Transactional` ensures all operations succeed together or all are rolled back.

## Running Locally

### Prerequisites

- Docker and Docker Compose
- Java 27 (for local development without Docker)

### With Docker Compose

```bash
docker-compose up --build
```

The application will be available at `http://localhost:8080`

## Core Concepts

### Without `@Transactional` (BAD)

```java
public void buyWithStatistics(Long productId) {
    product.buy();
    productRepository.save(product);        // Operation 1: SUCCEEDS

    // Exception thrown here - network error, database crash, etc.

    stats.incrementTotalSold();
    statsRepository.save(stats);      // Operation 2: NEVER EXECUTES
}
```

**Result:** Stock is reduced, but statistics not updated. Data is inconsistent.

### With `@Transactional` (GOOD)

```java
@Transactional
public void buyWithStatistics(Long productId) {
    product.buy();
    productRepository.save(product);        // Operation 1: SUCCEEDS

    // Exception thrown here - network error, database crash, etc.

    stats.incrementTotalSold();
    statsRepository.save(stats);      // Operation 2: NEVER EXECUTES
}
```

**Result:** BOTH operations rolled back. Product still has original stock. Data is consistent.

## Endpoints

The app is server-rendered: all routes return HTML or redirect back to `/products`.

- `GET /` or `GET /products` - List all products and display add form
- `POST /products` - Create new product
- `POST /products/{id}/buy` - Decrement stock and update statistics (`buyWithStatistics`)
- `POST /products/{fromId}/transfer/{toId}?stock=N` - Move stock between two products (`transferStock`, no form in the UI)

### Create Products

```bash
curl -X POST http://localhost:8080/products -d "name=Keyboard&description=Mechanical&price=99.90&stock=5"
curl -X POST http://localhost:8080/products -d "name=Mouse&description=Wireless&price=29.90&stock=5"
```

### Buy and Transfer

```bash
# Stock of product 1 goes down, total_sold in product_statistics goes up
curl -X POST http://localhost:8080/products/1/buy

# Moves 2 units from product 1 to product 2
curl -X POST "http://localhost:8080/products/1/transfer/2?stock=2"

# More than product 1 has: redirects to /products?error=insufficient-stock, nothing changes
curl -X POST "http://localhost:8080/products/1/transfer/2?stock=1000"
```

The transfer check runs before anything is modified. If a save failed after that point, `@Transactional` would roll both products back.

## Demonstrating Rollback

No endpoint triggers a failure. To see a rollback, edit `ProductService.buyWithStatistics` and enable the simulated failure between operation 1 and operation 2. Java rejects statements after a plain `throw` as unreachable, so write it with a condition:

```java
if (true) throw new RuntimeException("Simulated failure! Database connection lost!");
```

Rebuild with `docker compose up --build`, then:

```bash
# Returns HTTP 500
curl -i -X POST http://localhost:8080/products/1/buy
```

Stock of product 1 is unchanged and `product_statistics` has no new sale: both operations were rolled back.

To see the inconsistent case, also remove `@Transactional` from the method and replace `findByIdForUpdate` with `findById` (the locking query needs a running transaction). The same request then lowers the stock while the statistics stay behind.

## Key Architectural Decisions

### Service Layer Pattern

Transactions are managed at the service layer, not in controllers or repositories:

```
Controller (HTTP)
    ↓
Service (@Transactional)     ← Transaction boundary
    ↓           ↓
Repository  Repository
```

**Why?** Business operations often span multiple repository calls. The service layer is the natural place to define transaction boundaries.

### When to Use `@Transactional`

**Use it when:**
- Multiple database operations must succeed together
- Business logic spans multiple tables
- Updating related entities (product + statistics)

**Don't need it for:**
- Single read operation (`findAll`, `findById`)
- Single save operation (unless it updates multiple tables via relationships)

## Transaction Behavior

### Automatic Rollback

By default, `@Transactional` rolls back on:
- `RuntimeException` and its subclasses
- `Error` and its subclasses

Does NOT roll back on:
- Checked exceptions (unless explicitly configured)

### Custom Rollback Rules

```java
@Transactional(rollbackFor = {Exception.class})
public void conditionalRollback() {
    // Also rolls back on checked exceptions, in addition to the defaults above
}
```

## Database Schema

**products table:**
- `id` (bigserial, primary key)
- `name` (varchar)
- `description` (text)
- `price` (numeric(10,2), default 0)
- `stock` (integer, default 0)

**product_statistics table:**
- `id` (bigserial, primary key)
- `product_id` (bigint, unique)
- `total_sold` (integer, default 0)
- `last_updated` (timestamp)

## Proxy-Based Gotchas

Spring's `@Transactional` uses proxies, which means:

1. **Self-invocation doesn't work** - Method A calls method B in same class, B's `@Transactional` is ignored
2. **Must be public** - Private methods are not proxied
3. **Must be called externally** - Only works when called from outside the class

## Testing

Follow "Demonstrating Rollback" above and compare the tables after each step:

```
docker exec -it tx-java-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM products;"
docker exec -it tx-java-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM product_statistics;"

docker exec -it tx-java-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM flyway_schema_history;"
```

The console logs show the SQL statements (`spring.jpa.show-sql=true`).

`ProductServiceTest` and `ProductServiceIntegrationTest` (with Testcontainers) cover the same behaviour. See the comments at the end of `Dockerfile.test` for how to run them.

## Further Reading

- [Spring Transaction Management Documentation](https://docs.spring.io/spring-framework/reference/data-access/transaction.html)
- [Understanding @Transactional](https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/annotations.html)
- [Transaction Propagation](https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/tx-propagation.html)
