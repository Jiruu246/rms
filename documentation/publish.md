# Publish Workflow Documentation

Automated CI/CD pipeline for building and deploying RMS Docker containers.

## 🚀 Overview
The `Publish` workflow is responsible for containerizing the application services and pushing them to the **GitHub Container Registry (GHCR)**. This ensures that the `main` branch always has a corresponding "ready-to-deploy" image.

## 🛠 Workflow Configuration

### Triggers
- **Branch:** `main`
- **Event:** `push`

### Registry Details
- **Registry Host:** `ghcr.io`
- **Images Published:**
  1. `rms-api`: The core backend service.
  2. `rms-database-migrator`: The utility used for schema updates.

## 📦 Build Process

### 1. Environment Preparation
The workflow initializes by checking out the code and logging into GHCR using the `GITHUB_TOKEN`. It also normalizes the repository owner name to lowercase to prevent build failures due to Docker's naming constraints.

### 2. Versioning Strategy
We use a composite versioning string for traceability:
`YYYYMMDD-HHMM-shortSHA`

**Example:** `20240502-2015-7b3a1c2`

### 3. Build & Push
The workflow uses `docker/build-push-action@v5` to perform the following for both the API and the Migrator:
1. Build the image using the service-specific Dockerfile.
2. Tag the image with the unique **Version** tag.
3. Tag the image as **latest**.
4. Push both tags to GHCR.

## 🔑 Permissions
The workflow requires the following GITHUB_TOKEN permissions:
- `contents: read` (to access the code)
- `packages: write` (to upload images)
