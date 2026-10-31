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
src/main/java/ch/ost/i/dsl/ssr/
├── DemoApplication.java                    # Main application entry point
├── Product.java                               # JPA entity for products
├── ProductRepository.java                     # JPA repository for products
├── ProductStatistics.java                     # JPA entity for statistics
├── ProductStatisticsRepository.java           # JPA repository for statistics
├── ProductService.java                        # Service layer with @Transactional
├── ProductServiceWithFailureDemo.java         # Demo service with simulated failures
├── SimpleTransactionDemo.java              # Simplest possible demonstration
├── TransactionDemoController.java          # REST endpoints for testing
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

## Demo Endpoints

### Check Product State

```bash
curl http://localhost:8080/demo/product/1
```

Returns current product name and stock.

### Test WITH Transaction (Rollback Works)

```bash
# Check initial state
curl http://localhost:8080/demo/product/1
# Output: Stock: 5

# Trigger failure with @Transactional
curl -X POST http://localhost:8080/demo/explicit-failure/1
# Output: Exception thrown, rollback successful

# Verify rollback worked
curl http://localhost:8080/demo/product/1
# Output: Stock: 5 (unchanged - rollback worked!)
```

### Test WITHOUT Transaction (No Rollback - Data Corrupted)

```bash
# Check initial state
curl http://localhost:8080/demo/product/1
# Output: Stock: 5

# Trigger failure WITHOUT @Transactional
curl -X POST http://localhost:8080/demo/no-transaction/1
# Output: Exception thrown, NO ROLLBACK

# Verify no rollback (data corrupted)
curl http://localhost:8080/demo/product/1
# Output: Stock: 6 (increased - no rollback, data inconsistent!)
```

### Test Conditional Failure

```bash
# With failure flag (should rollback)
curl -X POST "http://localhost:8080/demo/conditional/1?fail=true"
# Output: Rollback successful, stock unchanged

# Without failure flag (should succeed)
curl -X POST "http://localhost:8080/demo/conditional/1?fail=false"
# Output: Success, stock incremented
```

### Test Business Rule Violation

```bash
# Transfer stock between products
# If target product would exceed 100 units, transfer fails and rolls back
curl -X POST "http://localhost:8080/demo/transfer?from=1&to=2&stock=10"
```

## Five Ways to Simulate Failures

The demo includes five different methods to simulate operation 2 failing:

1. **Explicit Exception** - Simply throw a `RuntimeException` (simplest)
2. **Database Constraint Violation** - Violate unique constraint (most realistic)
3. **Null Pointer Exception** - Simulate unexpected runtime error
4. **Conditional Failure** - Use boolean flag to control failure (best for testing)
5. **Business Rule Violation** - Validate business logic (e.g., max stock limit)

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
@Transactional(rollbackFor = {IllegalStateException.class})
public void conditionalRollback() {
    // Only rolls back on IllegalStateException
}
```

## Database Schema

**products table:**
- `id` (bigserial, primary key)
- `name` (varchar)
- `description` (text)
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

Run the demo endpoints in order:

1. Create a product with initial stock
2. Test with `@Transactional` - verify rollback works
3. Test without `@Transactional` - observe data corruption
4. Compare results

```
docker exec -it tx-java-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM products;"
docker exec -it tx-java-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM product_statistics;"

docker exec -it tx-java-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM flyway_schema_history;"
```

The console logs will show:
```
Operation 1 saved
Exception thrown: Simulated failure!
Transaction rolled back (or not, depending on @Transactional)
```

Check the database to verify whether the rollback occurred.

## Further Reading

- [Spring Transaction Management Documentation](https://docs.spring.io/spring-framework/reference/data-access/transaction.html)
- [Understanding @Transactional](https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/annotations.html)
- [Transaction Propagation](https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/tx-propagation.html)
