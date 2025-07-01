# Users Service

The Users Service is a central microservice responsible for user management, authorization, and security within the application ecosystem. 
Its primary goal is to provide a robust and secure foundation for all operations related to user identity.

## Table of Contents
- [Functionality Overview](#functionality-overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Post-Installation](#post-installation)
- [Troubleshooting](#troubleshooting)
- [Directory Structure](#directory-structure)
- [Testing](#testing)
- [Useful-commands](#useful-commands)
- [Useful-addresses](#useful-addresses)
- [License](#license)
- [Notes](#notes)

## Functionality Overview
The **Users Service** is a central microservice responsible for user management, authorization, and security within the application ecosystem. Its primary goal is to provide a robust and secure foundation for all operations related to user identity.

Key features of the service include:
- **User Lifecycle Management**: Full support for registration, account activation, verification, and account deletion processes.
- **Advanced Authorization**:
    - Generation and validation of secure JWTs (Ed25519).
    - Support for two-factor authentication (2FA).
    - Password reset mechanism.
- **Device Fingerprinting**: Unique identification of each device a user logs in from to enhance security and enable session management.
- **Access Control (Roles & Permissions)**: A system of roles and permissions that allows for fine-grained control over access to different parts of the system.
- **Event-Driven System**: Publication of key events (e.g., user registration, login) to a Kafka broker, enabling asynchronous communication between microservices.
- **Auditing and Activity Tracking**: Recording key actions performed by users.

## Prerequisites
- **Linux, Debian 12 preferred**: Windows k8s installation might be problematic for promtail configuration.
- **Golang, 1.24 or newer**: As dependencies in libraries required these versions.
- **Docker Engine 28.2.2 or newer**: Installed and running version.
- **Docker Desktop version 28.1.1 or newer**: If decided to using UI version.
- **Docker**: `Dockerfile` located at `./docker` for the main users service project folder.

## Installation
1. Clone the repository:
   ```bash
      git clone https://github.com/factory-chainline/users-service
      cd users-service
   ```
2. Downlaod the dependencies:
   ```bash
      go mod tidy
   ```

3. If the installation was problematic:
   - **Ensure you have Ent client downloaded properly**
   ```bash
      go get entgo.io/ent/cmd/ent 
   ```

   - **Get all of it's dependencies**
   ```bash
      go get entgo.io/ent/cmd/internal/printer@v0.14.4   
   ```

## Post-Installation
1. Work with code

## Troubleshooting
- **Continous Integration Actions Failed**:
  Ensure your code passed tests locally before commit. It will be checked anyway, but don't waste our common memory.
- **Docker not running**:
  Start Docker before running the tests.
- **Manifest Errors**:
  Ensure all `settings.yaml` in `tests/settings` folder are properly configurated for your local environment.

## Manually building, tagging and pushing docker images
- **Building docker images**
   ```bash
      docker build -t ghcr.io/factory-chainline/users-service:{current-version} -f ./docker/Dockerfile .   
   ``` 

- **Tagging docker image**
   ```bash
      
   ```   
- **Pushing docker image**
   ```bash
      
   ```  

## Directory Structure
```bash
.
.
.
├── app/
│   ├── backups/
│   ├── cmd/
│   ├── config/
│   ├── controllers/
│   ├── infrastructure/
│   ├── middlewares/
│   ├── models/
│   ├── repositories/
│   ├── services/
│   ├── services/
│   └── utils/
│   
├── tests/
│   ├── registry/
│   │   ├── config.go
│   │   ├── function_registry.go
│   │   ├── schema_registry.go
│   │   └── struct_registry.go
│   │
│   ├── resources/
│   │   ├── unit/*.json
│   │   ├── integration/*.json
│   │   └── e2e/*.json
│   │
│   ├── settings/
│   │   └── settings.yaml
│   │
│   ├── main_test.go
│   └── runner_test.go
│
└── Docker/
    ├── certificates/
    └── Dockerfile/

```
## Testing
- **Running unit tests**   
   ```bash
      go test -v -tags=unit ./tests -args -test_type=unit
   ```

- **Running integration test**   
   ```bash
      go test -v -tags=integration ./tests -args -test_type=integration
   ```

- **Running e2e test**   
   ```bash
      go test -v -tags=e2e ./tests -args -test_type=e2e
   ```

- **Test Setup JSON structures**
   ```bash
      
   ```

- **Test Step JSON structures**
   ```bash
      
   ```

## Useful commands
- **Generating keys for ED25519**
   Generate your private key:
   ```bash
      openssl genpkey -algorithm ED25519 -out private_key.pem
   ```
   Then resolve public key:
   ```bash
      openssl pkey -in private_key.pem -pubout -out public_key.pem
   ```

- **Generating swagger documentation**
   ```bash
      swag init -g ./app/cmd/main.go -o ./docker/docs/ --parseDependency --parseInternal --dir .     
   ``` 

- **Generating Ent schemas models**
   ```bash
      go run entgo.io/ent/cmd/ent generate ./app/models/ent/schema    
   ``` 

- **hosting documentation locally**
   ```bash
      
   ``` 

## Useful addresses
- **forwarding users-service port without ingress**   
   ```bash
      kubectl port-forward svc/users-service -n users-service 8000:8000  
   ```

- **swagger address**
   ```bash
      localhost:8000/swagger
   ```    


## License
- {TO DO}

## Notes
- {TO DO}

