## DevTest environment
This repo's test is ran from VSCode test runner
1. In order to run individual test from VSCode inside the test file directly, you need to put the .env.test within the /integration_test folder
- It's unclear how to point the env file to the root of the project in a clean manner so this is a compromise for now

2. In order to run and have an overview of all the test case from the Testing tab in VSCode you need to have a .env.test at the root level

These are the potential areas to be reconfigured when we have more resources

# Database Setup for Development Environment
You will run a docker instance of the PostgreSQL localy on your computer for development

Prerequisites
- Docker and Docker Compose installed
- Git repository cloned to your local machine

## Setup Instructions
1. Create environment file
```
cp .env.example .env
```

2. Configure database settings
Edit the `.env` file and update:

    - Database credentials (`POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`)
    - Data storage path (`POSTGRES_DATA_PATH`) - use forward slashes for Windows paths

3. Start the database
```
docker compose up -d db
```
4. Verify database is running
```
docker compose ps
```

5. Migrate
```
   go run ./cmd/migrate apply
```
For more migration details see [migration document](./migration-guide.md)

6. Seed
```
go run ./cmd/migrate seed
```

7. Run web server