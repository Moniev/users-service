# Users Service

## Table of Contents
- [Functionality Overview](#functionality-overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Post-Installation](#post-installation)
- [Troubleshooting](#troubleshooting)
- [Directory Structure](#directory-structure)
- [Testing](#testing)
- [Useful Commands](#useful-commands)
- [Useful Addresses](#useful-addresses)
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
- **Docker Desktop version 28.1.1 or newer**: If you decide to use the UI version.
- **Dockerfile**: Located at `./docker` for the main users service project folder.

## Installation
1. Clone the repository:
   ```bash
      git clone https://github.com/factory-chainline/users-service
      cd users-service
   ```
2. Download the dependencies:
   ```bash
      go mod tidy
   ```

3. If the installation was problematic:
   - **Ensure you have Ent client downloaded properly**
   ```bash
      go get entgo.io/ent/cmd/ent 
   ```

   - **Get all of its dependencies**
   ```bash
      go get entgo.io/ent/cmd/internal/printer@v0.14.4   
   ```

## Post-Installation
1. Work with the code.

## Troubleshooting
- **Continuous Integration Actions Failed**:
  Ensure your code passed tests locally before commit. It will be checked anyway, but don't waste our common memory.
- **Docker not running**:
  Start Docker before running the tests.
- **Manifest Errors**:
  Ensure all `settings.yaml` in the `tests/settings` folder are properly configured for your local environment.

## Manually building, tagging, and pushing Docker images
- **Building Docker images**
   ```bash
      docker build -t ghcr.io/factory-chainline/users-service:{current-version} -f ./docker/Dockerfile .   
   ```
   ```bash
      make build-docker 
   ``` 

- **Tagging a Docker image**
   ```bash
      docker tag ghcr.io/factory-chainline/users-service:<source> ghcr.io/factory-chainline/users-service:<target>  
   ```   
- **Pushing a Docker image**
   ```bash
      docker push ghcr.io/factory-chainline/users-service:<source>
   ```  

## Directory Structure
```bash
.
.
.
├── app
│   ├── cmd
│   ├── config
│   ├── controllers
│   ├── docs
│   ├── infrastructure
│   ├── middlewares
│   ├── models
│   │   ├── ent
│   │   │   ├── activationcode
│   │   │   ├── entrepreneurdetails
│   │   │   ├── enttest
│   │   │   ├── hook
│   │   │   ├── location
│   │   │   ├── migrate
│   │   │   ├── predicate
│   │   │   ├── resetcode
│   │   │   ├── rolepermission
│   │   │   ├── runtime
│   │   │   ├── schema
│   │   │   ├── secondfactorcode
│   │   │   ├── user
│   │   │   ├── useraction
│   │   │   ├── userdetails
│   │   │   ├── userdevice
│   │   │   ├── userrole
│   │   │   ├── usersettings
│   │   │   └── verificationcode
│   │   ├── events
│   │   ├── handlers
│   │   ├── requests
│   │   ├── responses
│   │   └── utils
│   ├── repositories
│   ├── routes
│   ├── services
│   └── utils
├── docker
└── tests
    ├── mocks
    ├── registry
    ├── resources
    │   ├── e2e
    │   ├── integration
    │   └── unit
    └── settings
```
## Testing
- **Running unit tests**
   ```bash
      make test-unit
   ```

- **Running integration tests**
   ```bash
      make test-integration
   ```

- **Running e2e tests**
   ```bash
      make test-e2e
   ```

### Defining Tests in JSON

The test runner is data-driven, meaning test cases are defined in `.json` files located in the `tests/resources/` directory. Each JSON file represents a test suite.

#### Test Setup (`setup_data`)

The `setup_data` array is used to populate the database with a specific state before a test case runs.

**Example Structure:**
```json
"setup_data": [
  {
    "model": "UserRole",
    "data": {
      "id": 1, 
      "name": "admin"
    }
  },
  {
    "model": "User",
    "data": {
      "id": 1,
      "mail": "admin@example.com",
      "password": "securepassword123",
      "user_role_ids": [1] 
    }
  }
]
```

**Key Conventions:**
- **`model`**: The name of the `ent` model to create (e.g., "User", "UserRole").
- **`data`**: A map of fields and their values for the new entity.
- **`id`**: A **local, temporary ID** used only within this file to define relationships. It does not correspond to the actual database ID.
- **Relationships (to-one)**: To link to a single entity, use the `_id` suffix (e.g., `"owner_id": 1`). This links to the entity that has the local ID of `1`.
- **Relationships (to-many)**: To link to multiple entities, use the `_ids` suffix (e.g., `"user_role_ids": [1, 2]`). This links to entities with local IDs `1` and `2`.
- **Order Matters**: Entities must be defined before they are referenced. For example, `UserRole` must be defined before a `User` can be assigned to it.

#### Test Steps (`steps`)

The `steps` array defines the sequence of actions and assertions for a test case.

**Example Structure:**
```json
{
   "setup_data": [
         {
            "model": "UserRole",
            "data": {
            "id": 1,
            "name": "user",
            "description": "Standard user role"
            }
         }
      ],
   "steps": [
      {
         "type": "executeFunction",
         "description": "Should return true for a valid email",
         "params": {
            "function_name": "utils.CheckEmailFormat",
            "args": ["test@example.com"]
         },
         "expected": {
            "return_value": true
         }
      },
      {
         "type": "httpRequest",
         "description": "Attempt to register a new user",
         "params": {
            "method": "POST",
            "path": "/api/v1/auth/register",
            "body": {
            "mail": "new.user@example.com",
            "password": "Password123!"
            }
         },
         "expected": {
            "statusCode": 200
         }
      }
   ]
}
```
**Key Fields:**
- **`type`**: The type of action to perform. Supported types are `executeFunction`, `httpRequest`, `dbRowCount`, and `dbAttributeEquals`.
- **`description`**: A human-readable description of the step.
- **`params`**: The parameters for the action (e.g., function name and arguments, or HTTP method and body).
- **`expected`**: The expected outcome (e.g., a specific return value or HTTP status code).
- **`store_result_as`**: (Optional) A key to store the result of this step, allowing it to be used as a dependency in subsequent steps (e.g., `{{myResult.data.id}}`).

## Useful Commands
- **Generating keys for ED25519**
   Generate your private key:
   ```bash
      openssl genpkey -algorithm ED25519 -out private_key.pem
   ```
   Then resolve the public key:
   ```bash
      openssl pkey -in private_key.pem -pubout -out public_key.pem
   ```

- **Generating Swagger documentation**
   ```bash
      swag init -g app/cmd/main.go -o app/docs   
   ``` 

- **Generating Ent schema models**
   ```bash
      go run entgo.io/ent/cmd/ent generate ./app/models/ent/schema    
   ``` 
- **Hosting documentation locally**
   ```bash
      godoc -http=:6060
   ``` 

## Useful Addresses
- **Forwarding users-service port without ingress**
   ```bash
      kubectl port-forward svc/users-service -n users-service 8000:8000  
   ```

- **Swagger address**
   ```bash
      localhost:8000/swagger
   ```  

- **Swagger address**
   ```bash
      localhost:8000/redoc
   ```    
## Updates
![Alt](https://repobeats.axiom.co/api/embed/0ccd32c70b973168f94e36384a64155cb9867f07.svg "Repobeats analytics image")
## License
© 2025 Robert Moń, All Rights Reserved. 
You may use it for noncommercial purpose if your name is not Kamil Rudyk or Mateusz Kacpura ;>
