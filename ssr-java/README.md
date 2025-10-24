# Spring Boot SSR Game Management Demo

A minimal Spring Boot application demonstrating server-side rendering with Thymeleaf, JPA, and PostgreSQL.

## Stack

- Java 25
- Spring Boot 3.5.6
- Thymeleaf (server-side templating)
- Spring Data JPA
- PostgreSQL 18
- Docker & Docker Compose

## Project Structure

```
src/main/java/ch/ost/i/dsl/ssr/
├── DemoApplication.java      # Main application entry point with sample data
├── Game.java                  # JPA entity
├── GameRepository.java        # JPA repository interface
└── GameController.java        # MVC controller

src/main/resources/
├── application.properties     # Minimal config
└── templates/
    └── games.html            # Thymeleaf template
```

## Running Locally

### Prerequisites

- Docker and Docker Compose
- Java 25 (for local development without Docker)

### With Docker Compose

```bash
docker-compose up --build
```

The application will be available at `http://localhost:8080`

## Features

- Add games with title and description
- List all games in a table
- Click star button to increment star count
- Server-side HTML rendering with Thymeleaf
- PostgreSQL persistence with automatic schema generation
- Sample data (Zelda, Mario) loaded on startup

## SSR Flow

1. Browser requests `GET /games`
2. Spring Boot controller queries PostgreSQL via JPA repository
3. Thymeleaf merges data with `games.html` template
4. Complete HTML with game data sent to browser
5. Browser displays content immediately (no JavaScript needed)
6. When user clicks star: `POST /games/{id}/star` → full page reload

## Configuration Notes

**Spring Boot "Magic"**

This project relies heavily on Spring Boot's convention-over-configuration:

- `spring-boot-starter-thymeleaf` → Auto-configures Thymeleaf, looks for templates in `src/main/resources/templates/`
- `spring-boot-starter-data-jpa` → Auto-implements repository methods, generates SQL
- `@SpringBootApplication` → Component scanning, auto-configuration
- JPA annotations → Automatic schema creation from entities
- Return `"games"` from controller → Resolves to `templates/games.html`

## Endpoints

- `GET /` or `GET /games` - List all games and display add form
- `POST /games` - Create new game
- `POST /games/{id}/star` - Increment star count (redirects back)

## Database

Schema is auto-generated from the `Game` entity on startup (`spring.jpa.hibernate.ddl-auto=update`).

Table: `games`
- `id` (bigserial, primary key)
- `title` (varchar)
- `description` (text)
- `stars` (integer, default 0)

Data persists in directory `.db` in the project directory.

## Comparison with CSR

**SSR (this app):**
- Server renders complete HTML with data
- Fast FCP, content visible immediately
- Full page reload on every interaction
- No JavaScript required for basic functionality

**CSR (Vue + Go app):**
- Server sends empty HTML shell
- JavaScript fetches data and renders UI
- No page reloads after initial load
- Requires JavaScript for all functionality