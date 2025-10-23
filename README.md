# Spring Boot SSR Demo

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
├── DemoApplication.java      # Main application entry point
├── User.java                  # JPA entity (uses Lombok)
├── UserRepository.java        # JPA repository interface
└── UserController.java        # MVC controller

src/main/resources/
├── application.properties     # Minimal config
└── templates/
    └── users.html            # Thymeleaf template
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

- CR(UD) operations on User entity
- Server-side HTML rendering with Thymeleaf
- PostgreSQL persistence with automatic schema generation
- Simple form handling and validation

## Configuration Notes

**Spring Boot "Magic"**

This project relies heavily on Spring Boot's convention-over-configuration:

- `spring-boot-starter-thymeleaf` → Auto-configures Thymeleaf, looks for templates in `src/main/resources/templates/`
- `spring-boot-starter-data-jpa` → Auto-implements repository methods, generates SQL
- `@SpringBootApplication` → Component scanning, auto-configuration
- JPA annotations → Automatic schema creation from entities
- Return `"users"` from controller → Resolves to `templates/users.html`

## Development

## Endpoints

- `GET /` or `GET /users` - List all users and display add form
- `POST /users` - Create new user

## Database

Schema is auto-generated from the `User` entity on startup (`spring.jpa.hibernate.ddl-auto=update`).

Table: `users`
- `id` (bigserial, primary key)
- `name` (varchar)
- `email` (varchar)

Data persists in Docker volume `postgres-data`.