# Database Migration Guide

This guide covers how to use Go Ent for database schema management and migrations in your development workflow.

## Overview

This project uses [Go Ent](https://entgo.io/) for ORM and database schema management with PostgreSQL. The migration system provides:

- **Schema-as-Code**: Define your database schema in Go code
- **Migrations Helper**: Generate and apply database changes using devtools
- **Branch-Safe Development**: Reset database when switching branches
- **Development Workflow**: Easy commands for common operations

## Development Workflows

### Scenario 1: Creating New Schema Changes

1. if you create a new table, then run the following to generate the schema file. or create a new file in the schema folder directly.
    ```go
    go run -mod=mod entgo.io/ent/cmd/ent new User
    ```

2. **Modify Schema**: Edit files in `internal/ent/schema/`
   ```go
   // internal/ent/schema/user.go
   func (User) Fields() []ent.Field {
       return []ent.Field{
           field.String("name").NotEmpty(),
           field.String("new_field").Optional(), // <- New field
           // ... other fields
       }
   }
   ```

3. **Regenerate Ent Code**:
   ```bash
   make ent-generate
   # or
   go generate ./internal/ent
   ```

4. **Apply Changes**:
   ```bash
   make migrate-up
   # or
   go run ./cmd/migraete apply
   ```

   This command will open a connection to the database, and execute the migration. It drops columns and indexes that are no longer defined in the schema and apply all pending schema changes to the database

### Scenario 2: Branch Switching (different Schema)

When switching to a branch with an older database schema:

```bash
# Switch to older branch
git checkout feature/older-schema

# run migrate apply to match current branch schema
go run ./cmd/migrate apply

```

### Scenario 4: Clean Development Start

For a fresh development environment:

```bash
# Get latest code
git pull origin main

# Clean slate database
go run ./cmd/migrate reset

```
### Migration Commands

| Command | Description | Usage |
|---------|-------------|--------|
| `apply` | Apply all pending migrations | `go run ./cmd/migrate apply` |
| `reset` | **DESTRUCTIVE** - Drop all tables and recreate | `go run ./cmd/migrate reset` |
| `create` | Generate migration SQL file | `go run ./cmd/migrate create add_user_table` |

# For production

**Never use apply to migrate**, instead generate the migration script and review before apply the migration

**Use descriptive migration names**:
   ```bash
   go run ./cmd/migrate create add_user_authentication
   go run ./cmd/migrate create update_category_constraints
   ```