# Spring Boot Transaction Management Demo

A minimal Spring Boot application demonstrating transaction management with `@Transactional`, showing the difference between atomic operations and non-atomic operations.

## Stack

- Java 25
- Spring Boot 3.5.6
- Spring Data JPA
- PostgreSQL 18
- Docker & Docker Compose

## Project Structure

```
src/main/java/ch/ost/i/dsl/ssr/
├── DemoApplication.java                    # Main application entry point
├── Game.java                               # JPA entity for games
├── GameRepository.java                     # JPA repository for games
├── GameStatistics.java                     # JPA entity for statistics
├── GameStatisticsRepository.java           # JPA repository for statistics
├── GameService.java                        # Service layer with @Transactional
├── GameServiceWithFailureDemo.java         # Demo service with simulated failures
├── SimpleTransactionDemo.java              # Simplest possible demonstration
├── TransactionDemoController.java          # REST endpoints for testing
└── GameController.java                     # MVC controller
```

## What This Demo Shows

**The Problem:** When multiple database operations must succeed together (atomicity), exceptions can leave data in an inconsistent state.

**The Solution:** `@Transactional` ensures all operations succeed together or all are rolled back.

## Running Locally

### Prerequisites

- Docker and Docker Compose
- Java 25 (for local development without Docker)

### With Docker Compose

```bash
docker-compose up --build
```

The application will be available at `http://localhost:8080`

## Core Concepts

### Without `@Transactional` (BAD)

```java
public void addStarWithStatistics(Long gameId) {
    game.addStar();
    gameRepository.save(game);        // Operation 1: SUCCEEDS

    // Exception thrown here - network error, database crash, etc.

    stats.incrementTotalStars();
    statsRepository.save(stats);      // Operation 2: NEVER EXECUTES
}
```

**Result:** Game has +1 star, but statistics not updated. Data is inconsistent.

### With `@Transactional` (GOOD)

```java
@Transactional
public void addStarWithStatistics(Long gameId) {
    game.addStar();
    gameRepository.save(game);        // Operation 1: SUCCEEDS

    // Exception thrown here - network error, database crash, etc.

    stats.incrementTotalStars();
    statsRepository.save(stats);      // Operation 2: NEVER EXECUTES
}
```

**Result:** BOTH operations rolled back. Game still has original star count. Data is consistent.

## Demo Endpoints

### Check Game State

```bash
curl http://localhost:8080/demo/game/1
```

Returns current game title and star count.

### Test WITH Transaction (Rollback Works)

```bash
# Check initial state
curl http://localhost:8080/demo/game/1
# Output: Stars: 5

# Trigger failure with @Transactional
curl -X POST http://localhost:8080/demo/explicit-failure/1
# Output: Exception thrown, rollback successful

# Verify rollback worked
curl http://localhost:8080/demo/game/1
# Output: Stars: 5 (unchanged - rollback worked!)
```

### Test WITHOUT Transaction (No Rollback - Data Corrupted)

```bash
# Check initial state
curl http://localhost:8080/demo/game/1
# Output: Stars: 5

# Trigger failure WITHOUT @Transactional
curl -X POST http://localhost:8080/demo/no-transaction/1
# Output: Exception thrown, NO ROLLBACK

# Verify no rollback (data corrupted)
curl http://localhost:8080/demo/game/1
# Output: Stars: 6 (increased - no rollback, data inconsistent!)
```

### Test Conditional Failure

```bash
# With failure flag (should rollback)
curl -X POST "http://localhost:8080/demo/conditional/1?fail=true"
# Output: Rollback successful, stars unchanged

# Without failure flag (should succeed)
curl -X POST "http://localhost:8080/demo/conditional/1?fail=false"
# Output: Success, stars incremented
```

### Test Business Rule Violation

```bash
# Transfer stars between games
# If target game would exceed 100 stars, transfer fails and rolls back
curl -X POST "http://localhost:8080/demo/transfer?from=1&to=2&stars=10"
```

## Five Ways to Simulate Failures

The demo includes five different methods to simulate operation 2 failing:

1. **Explicit Exception** - Simply throw a `RuntimeException` (simplest)
2. **Database Constraint Violation** - Violate unique constraint (most realistic)
3. **Null Pointer Exception** - Simulate unexpected runtime error
4. **Conditional Failure** - Use boolean flag to control failure (best for testing)
5. **Business Rule Violation** - Validate business logic (e.g., max stars limit)

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
- Updating related entities (game + statistics)

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

**games table:**
- `id` (bigserial, primary key)
- `title` (varchar)
- `description` (text)
- `stars` (integer, default 0)

**game_statistics table:**
- `id` (bigserial, primary key)
- `game_id` (bigint, unique)
- `total_stars` (integer, default 0)
- `last_updated` (timestamp)

## Proxy-Based Gotchas

Spring's `@Transactional` uses proxies, which means:

1. **Self-invocation doesn't work** - Method A calls method B in same class, B's `@Transactional` is ignored
2. **Must be public** - Private methods are not proxied
3. **Must be called externally** - Only works when called from outside the class

## Testing

Run the demo endpoints in order:

1. Create a game with initial stars
2. Test with `@Transactional` - verify rollback works
3. Test without `@Transactional` - observe data corruption
4. Compare results

```
docker exec -it tx-java-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM games;"
docker exec -it tx-java-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM game_statistics;"

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
