# Go Transaction Management Demo

A minimal Go application demonstrating explicit transaction management with `Begin()`, `Commit()`, and `Rollback()`, showing the difference between atomic operations and non-atomic operations.

## Stack

- Go 1.25+
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
- Go 1.25+ (for local development without Docker)

### With Docker Compose

```bash
docker compose up --build
```

The application will be available at `http://localhost:8081`

## Core Concepts

### Without Transaction (BAD)

```go
func addStarWithoutTransaction(gameID int) error {
    // Operation 1: Update game (NO TRANSACTION)
    _, err := db.Exec("UPDATE games SET stars = stars + 1 WHERE id = $1", gameID)
    if err != nil {
        return err
    }
    // Stars are ALREADY SAVED to database!
    
    // Failure happens here
    return errors.New("Network error!")
    
    // Operation 2 never executes
    // Result: Game has +1 star, statistics not updated
}
```

**Result:** Game has +1 star, but statistics not updated. Data is inconsistent.

### With Transaction (GOOD)

```go
func addStarWithTransaction(gameID int) error {
    // Begin transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()  // Rollback if commit not called
    
    // Operation 1: Update game
    _, err = tx.Exec("UPDATE games SET stars = stars + 1 WHERE id = $1", gameID)
    if err != nil {
        return err  // Rollback happens via defer
    }
    
    // Failure happens here
    return errors.New("Network error!")  // Triggers rollback
    
    // Operation 2 never executes
    
    // Commit would go here (never reached)
    // return tx.Commit()
}
```

**Result:** BOTH operations rolled back. Game still has original star count. Data is consistent.

## API Endpoints

### Basic CRUD Endpoints

- `GET /api/games` - List all games
- `POST /api/games` - Create new game
- `POST /api/games/{id}/star` - Add star (simple, no transaction demo)

### Transaction Demo Endpoints

#### Check Game State

```bash
curl http://localhost:8080/api/demo/game/1
```

Returns current game details including star count.

#### Test WITH Transaction (Rollback Works)

```bash
# Check initial state
curl http://localhost:8081/api/demo/game/1
# Output: {"id":1,"title":"The Legend of Zelda","stars":5}

# Trigger operation with transaction
curl -X POST http://localhost:8081/api/demo/with-transaction/1

# Response shows rollback worked:
# {
#   "error": "Simulated failure!",
#   "stars_before": 5,
#   "stars_after": 5,
#   "rolled_back": true,
#   "message": "Transaction rolled back! Stars unchanged."
# }
```

#### Test WITHOUT Transaction (No Rollback - Data Corrupted)

```bash
# Check initial state
curl http://localhost:8081/api/demo/game/1
# Output: {"stars":5}

# Trigger operation WITHOUT transaction
curl -X POST http://localhost:8081/api/demo/without-transaction/1

# Response shows NO rollback:
# {
#   "error": "Simulated failure!",
#   "stars_before": 5,
#   "stars_after": 6,
#   "rolled_back": false,
#   "message": "NO ROLLBACK! Stars increased despite error"
# }
```

#### Test Transfer Stars

```bash
# Transfer 10 stars from game 1 to game 2
curl -X POST http://localhost:8081/api/demo/transfer \
  -H "Content-Type: application/json" \
  -d '{"from_id":1,"to_id":2,"stars":10}'

# If transfer would make game 2 exceed 100 stars:
# {
#   "error": "Target game would exceed 100 stars",
#   "from_game": {"before": 50, "after": 50},
#   "to_game": {"before": 95, "after": 95},
#   "rolled_back": true,
#   "message": "Transaction rolled back! Both games unchanged."
# }

# Successful transfer:
# {
#   "success": true,
#   "from_game": {"before": 50, "after": 40},
#   "to_game": {"before": 30, "after": 40},
#   "message": "Transfer successful!"
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
public void addStar(Long gameId) {
    game.addStar();
    gameRepository.save(game);  // Transaction automatic
    
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
5. defer tx.Rollback()          → Never executes (commit called)
```

### Failed Transaction

```
1. tx, err := db.Begin()        → Start transaction
2. tx.Exec(...)                 → Operation 1 succeeds
3. return error                 → Operation 2 fails
4. defer tx.Rollback()          → ROLLBACK (data discarded)
5. tx.Commit()                  → Never reached
```

## Database Schema

**games table:**
- `id` (serial, primary key)
- `title` (varchar)
- `description` (text)
- `stars` (integer, default 0)

**game_statistics table:**
- `id` (serial, primary key)
- `game_id` (integer, unique, foreign key)
- `total_stars` (integer, default 0)
- `last_updated` (timestamp)

## Frontend Features

The web UI provides interactive transaction testing:

### Basic Features
- List all games in a table
- Add new games via form
- Click star button to increment count
- Client-side rendering with Vue.js

### Transaction Demo Features
- **With TX button** - Tests operation with transaction, shows rollback on failure
- **Without TX button** - Tests operation without transaction, demonstrates data corruption
- **Transfer form** - Transfer stars between games with business rule validation
- **Result display** - Shows JSON response with before/after state

### Using the Frontend

1. Open `http://localhost:8081` in browser
2. See list of sample games (Zelda, Mario, Metroid)
3. Click "With TX" button on any game
   - Observe: Stars remain unchanged (rollback worked)
   - Check result box: Shows `rolled_back: true`
4. Click "Without TX" button (with confirmation)
   - WARNING: This demonstrates data corruption
   - Observe: Stars increased despite error
   - Check result box: Shows `rolled_back: false`
5. Use transfer form to move stars between games
   - If target would exceed 100 stars, transfer rolls back
   - Both games remain unchanged

## Testing the Demo

### Manual Testing via curl

1. Check initial state:
```bash
curl http://localhost:8081/api/demo/game/1
```

2. Test WITH transaction (should rollback):
```bash
curl -X POST http://localhost:8081/api/demo/with-transaction/1
```

3. Verify rollback worked:
```bash
curl http://localhost:8081/api/demo/game/1
# Star count should be UNCHANGED
```

4. Test WITHOUT transaction (no rollback):
```bash
curl -X POST http://localhost:8081/api/demo/without-transaction/1
```

5. Verify no rollback (data corrupted):
```bash
curl http://localhost:8081/api/demo/game/1
# Star count should be INCREASED (inconsistent!)
```

### Verify in Database

```bash
# Connect to PostgreSQL
docker compose exec db psql -U postgres -d gamedb

# Check games table
SELECT * FROM games;

# Check statistics table
SELECT * FROM game_statistics;

# Join to see relationship
SELECT g.id, g.title, g.stars, gs.total_stars, gs.last_updated
FROM games g 
LEFT JOIN game_statistics gs ON g.id = gs.game_id;
```

## Best Practices in Go

### 1. Always Use defer for Rollback

```go
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()  // Safety net - always include this

// ... operations ...

return tx.Commit()  // If commit succeeds, rollback is no-op
```

**Why:** The `defer` ensures rollback happens even if code panics or returns early.

### 2. Pass Transaction Explicitly

```go
func updateGame(tx *sql.Tx, gameID int) error {
    // Function receives tx as parameter
    _, err := tx.Exec("UPDATE games SET stars = stars + 1 WHERE id = $1", gameID)
    return err
}

func updateStats(tx *sql.Tx, gameID int) error {
    // Same tx used across functions
    _, err := tx.Exec("UPDATE game_statistics ...")
    return err
}
```

**Why:** Makes transaction scope explicit and visible.

### 3. Return Errors, Let Caller Handle Rollback

```go
func businessLogic(gameID int) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    if err := updateGame(tx, gameID); err != nil {
        return err  // Rollback via defer
    }
    
    if err := updateStats(tx, gameID); err != nil {
        return err  // Rollback via defer
    }
    
    return tx.Commit()
}
```

**Why:** Separation of concerns - functions focus on logic, defer handles cleanup.

## Common Pitfalls

### Pitfall 1: Forgetting to use `tx`

```go
tx, _ := db.Begin()
defer tx.Rollback()

db.Exec("UPDATE ...")  // WRONG! Using db instead of tx
// This executes OUTSIDE the transaction!

tx.Exec("UPDATE ...")  // CORRECT! Using tx
```

### Pitfall 2: Not Using defer

```go
tx, _ := db.Begin()
// WRONG! No defer tx.Rollback()

if err := doSomething(tx); err != nil {
    return err  // Transaction never rolled back - connection leak!
}

// CORRECT:
tx, _ := db.Begin()
defer tx.Rollback()
```

### Pitfall 3: Ignoring Errors

```go
// WRONG
tx, _ := db.Begin()  // Ignoring error

// CORRECT
tx, err := db.Begin()
if err != nil {
    return err
}
```

## Console Output

Watch the server logs to see transaction behavior:

**With Transaction:**
```
Operation 1: Game stars updated
Simulated failure! Network error!
Transaction rolled back (via defer)
```

**Without Transaction:**
```
Operation 1: Game stars updated (SAVED TO DB)
Simulated failure! But operation 1 already committed!
NO ROLLBACK - data is inconsistent!
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
- Updating related entities (game + statistics)

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