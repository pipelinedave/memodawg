# MemoDawg Transcription Service

## Introduction

MemoDawg is a transcription service that utilizes Microsoft Azure's Speech-to-Text API to transcribe audio files. This project consists of a frontend and a backend API, both containerized using Docker, and orchestrated via Kubernetes.

## Table of Contents

- [Getting Started](#getting-started)
- [Installation](#installation)
- [Project Structure](#project-structure)
- [API](#api)
- [Frontend](#frontend)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [License](#license)

## Getting Started

### Prerequisites

- Docker
- Kubernetes
- Helm (Optional)
- Microsoft Azure account

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/pipelinedave/MemoDawg.git
    ```

2. Navigate to the backend folder
   ```bash
   cd memodawg/backend
   ```

3. Create a `.env` file and fill it with the necessary environment variables:

   ```text
   AZURE_KEY=your-azure-key
   GOTIFY_TOKEN=your-gotify-token
   MEMODAWG_KEY=your-memodawg-key
   ```

4. Build and run the Docker container:
   ```bash
   docker-compose up --build
   ```

#### Testing

Run tests using:

```bash
go test ./...

### Frontend

#### Local Development

1. Navigate to the frontend folder
   ```bash
   cd memodawg/frontend
   ```

2. Build and run the Docker container:
   ```bash
   docker-compose up --build
   ```

#### Testing

Run tests using:
```bash
go test ./...
```

---

## Deployment

Deployment is managed using Kubernetes. Please ensure you have a running Kubernetes cluster and `kubectl` configured before proceeding.

1. Create a namespace for the MemoDawg service:

   ```bash
   kubectl create namespace memodawg
   ```

2. Apply Kubernetes manifests:

   ```bash
   kubectl apply -f kubernetes/
   ```

This will create the necessary deployments, services, config maps, and secrets.

For updating the service, you can use:

```bash
kubectl rollout restart deployment memodawg-api -n memodawg
kubectl rollout restart deployment memodawg-frontend -n memodawg
```

## Contributing

Feel free to open issues or pull requests to improve the project. All contributions are welcome.

## License

This project is licensed under the MIT License. See `LICENSE` for details.
```

You can append this content to the part that you already copied. If you have further requirements, let me know!
