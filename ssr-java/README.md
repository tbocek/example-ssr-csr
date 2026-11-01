# Spring Boot SSR Shop Demo

A minimal Spring Boot application demonstrating server-side rendering with Thymeleaf, JPA, and PostgreSQL.

## Stack

- Java 27
- Spring Boot 4.1.2
- Thymeleaf (server-side templating)
- Spring Data JPA
- PostgreSQL 18
- Docker & Docker Compose

## Project Structure

```
src/main/java/ch/ost/i/dsl/ssr/
├── DemoApplication.java      # Main application entry point
├── Product.java               # JPA entity
├── ProductRepository.java     # JPA repository interface
└── ProductController.java     # MVC controller

src/main/resources/
├── application.properties     # Minimal config
└── templates/
    └── products.html         # Thymeleaf template
```

## Running Locally

### Prerequisites

- Docker and Docker Compose
- Java 27 and Gradle 9.8 or later (for local development without Docker; Gradle 9.8 is the first release that runs on Java 27)

### With Docker Compose

```bash
docker-compose up --build
```

The application will be available at `http://localhost:8081`

## Features

- Add products with name, description, price and stock
- List all products in a table
- Click buy button to decrement stock (disabled when sold out)
- Server-side HTML rendering with Thymeleaf
- PostgreSQL persistence with automatic schema generation

## SSR Flow

1. Browser requests `GET /products`
2. Spring Boot controller queries PostgreSQL via JPA repository
3. Thymeleaf merges data with `products.html` template
4. Complete HTML with product data sent to browser
5. Browser displays content immediately (no JavaScript needed)
6. When user clicks buy: `POST /products/{id}/buy` → full page reload

## Configuration Notes

**Spring Boot "Magic"**

This project relies heavily on Spring Boot's convention-over-configuration:

- `spring-boot-starter-thymeleaf` → Auto-configures Thymeleaf, looks for templates in `src/main/resources/templates/`
- `spring-boot-starter-data-jpa` → Auto-implements repository methods, generates SQL
- `@SpringBootApplication` → Component scanning, auto-configuration
- JPA annotations → Automatic schema creation from entities
- Return `"products"` from controller → Resolves to `templates/products.html`

## Endpoints

- `GET /` or `GET /products` - List all products and display add form
- `POST /products` - Create new product
- `POST /products/{id}/buy` - Decrement stock (redirects back)

## Database

Schema is auto-generated from the `Product` entity on startup (`spring.jpa.hibernate.ddl-auto=update`).

Table: `products`
- `id` (bigserial, primary key)
- `name` (varchar)
- `description` (text)
- `price` (numeric(10,2))
- `stock` (integer)

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