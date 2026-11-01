# Go Transaction Management Demo

A minimal Go application demonstrating explicit transaction management with `Begin()`, `Commit()`, and `Rollback()`, showing the difference between atomic operations and non-atomic operations.

## Stack

- Go 1.27+
- PostgreSQL 18
- Pure `database/sql` (no ORM)
- Vue.js 3 (frontend)
- Docker & Docker Compose

## What This Demo Shows

**The Problem:** When multiple database operations must succeed together (atomicity), exceptions can leave data in an inconsistent state.

**The Solution:** Explicit transaction management with `tx.Begin()`, `tx.Commit()`, and `tx.Rollback()` ensures all operations succeed together or all are rolled back.

## Running Locally

### Prerequisites

- Docker and Docker Compose
- Go 1.27+ (for local development without Docker)

### With Docker Compose

```bash
docker compose up --build
```

The application will be available at `http://localhost:8081`

## Core Concepts

### Without Transaction (BAD)

```go
func transferStockWithoutTransaction(fromID, toID, stock int) error {
    // Operation 1: Deduct stock from source product (NO TRANSACTION)
    _, err := db.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", stock, fromID)
    if err != nil {
        return err
    }
    // Stock is ALREADY SAVED to database!

    // Simulated failure before operation 2
    return fmt.Errorf("network error before operation 2")

    // Operation 2 never executes
}
```

**Result:** Stock was deducted from the source but never added to the target. Data is inconsistent.

### With Transaction (GOOD)

```go
func moveStockWithTransaction(productID int) error {
    const fromProductID = 1  // Product 1 receives the unit
    toProductID := productID // Target product loses the unit

    // Begin transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()  // No effect once commit has succeeded

    // Operation 1: Deduct unit from target product
    _, err = tx.Exec("UPDATE products SET stock = stock - 1 WHERE id = $1", toProductID)
    if err != nil {
        return err  // Rollback happens via defer
    }

    // Uncomment to test rollback:
    // return fmt.Errorf("Simulated failure! Network error!")

    // Operation 2: Add unit to product 1
    _, err = tx.Exec("UPDATE products SET stock = stock + 1 WHERE id = $1", fromProductID)
    if err != nil {
        return err  // Rollback happens via defer
    }

    return tx.Commit()
}
```

**Result:** Both operations commit, or neither does. With the failure line uncommented, operation 1 is rolled back and the product keeps its original stock.

## API Endpoints

### Basic CRUD Endpoints

- `GET /api/products` - List all products
- `POST /api/products` - Create new product
- `POST /api/products/{id}/buy` - Decrement stock (simple, no transaction demo)

### Transaction Demo Endpoints

Create products first (see `POST /api/products` above or use the frontend): the database starts empty.

#### Check Product State

```bash
curl http://localhost:8081/api/demo/product/1
```

Returns current product details including stock.

#### Test WITH Transaction

Moves one unit from product `{id}` to product 1, so use an ID other than 1 to see a change.

```bash
# Check initial state
curl http://localhost:8081/api/demo/product/2
# Output: {"id":2,"name":"Mechanical Keyboard","description":"...","price":99.9,"stock":5}

# Trigger operation with transaction
curl -X POST http://localhost:8081/api/demo/with-transaction/2

# As shipped, the transaction commits:
# {
#   "success": true,
#   "stock_before": 5,
#   "stock_after": 4,
#   "message": "Transaction committed successfully!"
# }
```

To see the rollback, uncomment the `return fmt.Errorf("Simulated failure! Network error!")` line in `moveStockWithTransaction` and rebuild with `docker compose up --build`:

```bash
# {
#   "error": "Simulated failure! Network error!",
#   "stock_before": 5,
#   "stock_after": 5,
#   "rolled_back": true,
#   "message": "Transaction rolled back! Stock unchanged."
# }
```

#### Test WITHOUT Transaction

`moveStockWithoutTransaction` is a single statement with no simulated failure: it adds one unit to product 1, whatever `{id}` is, and reports product 1's stock.

```bash
curl -X POST http://localhost:8081/api/demo/without-transaction/1

# {
#   "success": true,
#   "stock_before": 5,
#   "stock_after": 6,
#   "message": "Unit added to product 1 successfully (simple operation, no transaction needed)"
# }
```

#### Test Transfer Stock

```bash
# Transfer 10 units from product 1 to product 2
curl -X POST http://localhost:8081/api/demo/transfer \
  -H "Content-Type: application/json" \
  -d '{"from_id":1,"to_id":2,"stock":10}'

# If transfer would make product 2 exceed 100 units:
# {
#   "error": "target product would exceed 100 units (95 + 10 = 105)",
#   "from_product": {"before": 50, "after": 50},
#   "to_product": {"before": 95, "after": 95},
#   "rolled_back": true,
#   "message": "Transaction rolled back! Both products unchanged."
# }

# Successful transfer:
# {
#   "success": true,
#   "from_product": {"before": 50, "after": 40},
#   "to_product": {"before": 30, "after": 40},
#   "message": "Transfer successful!"
# }
```

#### Test Transfer WITHOUT Transaction (No Rollback - Data Corrupted)

This endpoint always fails between the two operations.

```bash
curl -X POST http://localhost:8081/api/demo/transfer-no-tx \
  -H "Content-Type: application/json" \
  -d '{"from_id":1,"to_id":2,"stock":10}'

# {
#   "error": "network error before operation 2",
#   "from_product": {"before": 50, "after": 40},
#   "to_product": {"before": 30, "after": 30},
#   "rolled_back": false,
#   "message": "NO ROLLBACK! First operation committed, second operation never executed. Data is inconsistent!"
# }
```

## Key Differences: Go vs Spring Boot

### Go (Explicit)

```go
tx, err := db.Begin()           // Explicit start
defer tx.Rollback()             // Explicit rollback (safety net)

_, err = tx.Exec("UPDATE ...")  // Use tx, not db
if err != nil {
    return err                  // Rollback via defer
}

return tx.Commit()              // Explicit commit
```

**Characteristics:**
- Manual `Begin()`, `Commit()`, `Rollback()`
- Transaction object (`tx`) passed explicitly
- `defer tx.Rollback()` pattern for safety
- Every step is visible in code
- No surprises - what you see is what you get
- Compiler enforces correct usage

### Spring Boot (Declarative)

```java
@Transactional
public void buy(Long productId) {
    product.buy();
    productRepository.save(product);  // Transaction automatic
    
    // Exception triggers automatic rollback
}
```

**Characteristics:**
- Annotation-based (`@Transactional`)
- Framework manages transaction lifecycle
- Automatic rollback on exceptions
- Transaction boundaries hidden
- Proxy-based (gotchas possible)
- Runtime errors for misconfiguration

## Transaction Flow in Go

### Successful Transaction

```
1. tx, err := db.Begin()        → Start transaction
2. tx.Exec(...)                 → Operation 1
3. tx.Exec(...)                 → Operation 2  
4. tx.Commit()                  → Commit (data saved)
5. defer tx.Rollback()          → Runs, but has no effect (already committed)
```

### Failed Transaction

```
1. tx, err := db.Begin()        → Start transaction
2. tx.Exec(...)                 → Operation 1 succeeds
3. return error                 → Operation 2 fails
4. defer tx.Rollback()          → ROLLBACK (data discarded)
5. tx.Commit()                  → Never reached
```

### Using the Frontend

1. Open `http://localhost:8081` in browser
2. Add a few products using the form
3. Enter two product IDs and a number of units in the "Transfer Stock" form
4. Click "Transfer WITH Transaction"
   - If target would exceed 100 units, transfer rolls back
   - Both products remain unchanged
5. Click "Transfer WITHOUT Transaction"
   - WARNING: This demonstrates data corruption
   - Observe: Stock deducted from source, never added to target
   - Check result box: Shows `rolled_back: false`

### Verify in Database

```
docker exec -it tx-go-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM products;"
docker exec -it tx-go-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM product_statistics;"


docker exec -it tx-go-postgres-1 psql -U postgres -d postgres -c "SELECT * FROM schema_migrations;"

go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
#does not crash as product_statistics here not used...
#port 5432 is not published by docker-compose.yml, add it to the postgres service first
migrate -path backend/migrations -database "postgres://postgres:password@localhost:5432/postgres?sslmode=disable" down 1
```

## Architecture Comparison

### Go Philosophy
- **Explicit over implicit** - Every step is visible
- **No magic** - No hidden framework behavior
- **Compile-time safety** - Type system catches errors early
- **Verbose but clear** - More code, but easier to understand

### Spring Boot Philosophy
- **Convention over configuration** - Framework handles wiring
- **Declarative** - Annotations describe intent
- **Runtime flexibility** - Dynamic proxy-based behavior
- **Concise but hidden** - Less code, but harder to trace

## When to Use Transactions

**Use transactions when:**
- Multiple database operations must succeed together
- Business logic spans multiple tables
- Data consistency is critical (money, inventory, user accounts)
- Updating related entities (product + statistics)

**Don't need transactions for:**
- Single read operation (`SELECT`)
- Single insert/update that doesn't depend on other data
- Operations where eventual consistency is acceptable
- Read-only queries

## Performance Considerations

**Transactions have overhead:**
- Lock acquisition and management
- Additional database round trips
- Connection pool pressure with long transactions

**Best practices:**
- Keep transactions short
- Perform non-database work outside transactions
- Consider isolation levels for concurrent access
- Use connection pooling appropriately

## Further Reading

- [Go database/sql Tutorial](https://go.dev/doc/database/execute-transactions)
- [PostgreSQL Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
- [Effective Go - Defer](https://go.dev/doc/effective_go#defer)
- [ACID Properties](https://en.wikipedia.org/wiki/ACID)