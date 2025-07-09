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

- **Tagging a Docker image**
   ```bash
      
   ```   
- **Pushing a Docker image**
   ```bash
      
   ```  

## Directory Structure
```bash
📦users-service
 ┣ 📂.github
 ┃ ┗ 📂workflows
 ┃ ┃ ┣ 📜dependabot-autocommit.yaml
 ┃ ┃ ┣ 📜main-ci-cd.yaml
 ┃ ┃ ┣ 📜reusable-build-push.yaml
 ┃ ┃ ┣ 📜reusable-e2e-tests.yaml
 ┃ ┃ ┣ 📜reusable-integration-tests.yaml
 ┃ ┃ ┣ 📜reusable-notify-discord.yaml
 ┃ ┃ ┣ 📜reusable-unit-tests.yaml
 ┃ ┃ ┣ 📜reusable-update-gitops.yaml
 ┃ ┃ ┗ 📜reusable-update-requirements.yaml
 ┣ 📂app
 ┃ ┣ 📂backups
 ┃ ┣ 📂cmd
 ┃ ┃ ┗ 📜main.go
 ┃ ┣ 📂config
 ┃ ┃ ┣ 📜config.go
 ┃ ┃ ┣ 📜database.go
 ┃ ┃ ┣ 📜kafka.go
 ┃ ┃ ┣ 📜logger.go
 ┃ ┃ ┣ 📜redis.go
 ┃ ┃ ┗ 📜settings.go
 ┃ ┣ 📂controllers
 ┃ ┃ ┣ 📜auth_controller.go
 ┃ ┃ ┣ 📜diagnostics_controller.go
 ┃ ┃ ┣ 📜docs_controller.go
 ┃ ┃ ┣ 📜handler.go
 ┃ ┃ ┗ 📜users_controller.go
 ┃ ┣ 📂infrastructure
 ┃ ┃ ┣ 📜cache_store.go
 ┃ ┃ ┣ 📜event_listener.go
 ┃ ┃ ┗ 📜event_notifier.go
 ┃ ┣ 📂middlewares
 ┃ ┃ ┣ 📜auth_middleware.go
 ┃ ┃ ┣ 📜config.go
 ┃ ┃ ┣ 📜kafka_message_handler_middleware.go
 ┃ ┃ ┣ 📜limit_requests_middleware.go
 ┃ ┃ ┣ 📜logger_middleware.go
 ┃ ┃ ┣ 📜metrics.go
 ┃ ┃ ┣ 📜metrics_cors_middleware.go
 ┃ ┃ ┣ 📜prometheus_middleware.go
 ┃ ┃ ┗ 📜request_identification_middleware.go
 ┃ ┣ 📂models
 ┃ ┃ ┃ 📂ent     
 ┃ ┃ ┃ ┣ 📂schema
 ┃ ┃ ┃ ┃ ┣ 📜activation_code.go
 ┃ ┃ ┃ ┃ ┣ 📜blacklisted_token.go
 ┃ ┃ ┃ ┃ ┣ 📜reset_code.go
 ┃ ┃ ┃ ┃ ┣ 📜role_permission.go
 ┃ ┃ ┃ ┃ ┣ 📜second_factor_code.go
 ┃ ┃ ┃ ┃ ┣ 📜user.go
 ┃ ┃ ┃ ┃ ┣ 📜user_action.go
 ┃ ┃ ┃ ┃ ┣ 📜user_ban.go
 ┃ ┃ ┃ ┃ ┣ 📜user_details.go
 ┃ ┃ ┃ ┃ ┣ 📜user_device.go
 ┃ ┃ ┃ ┃ ┣ 📜user_role.go
 ┃ ┃ ┃ ┃ ┣ 📜user_settings.go
 ┃ ┃ ┃ ┃ ┗ 📜verification_code.go
 ┃ ┃ ┣ 📂events
 ┃ ┃ ┃ ┗ 📜events.go
 ┃ ┃ ┣ 📂requests
 ┃ ┃ ┃ ┣ 📜auth.go
 ┃ ┃ ┃ ┣ 📜diagnostics.go
 ┃ ┃ ┃ ┣ 📜users.go
 ┃ ┃ ┃ ┗ 📜web_socket_helper.go
 ┃ ┃ ┃ 📂responses
 ┃ ┃ ┃ ┣ 📜error.go
 ┃ ┃ ┃ ┣ 📜messages.go
 ┃ ┃ ┃ ┣ 📜reporter.go
 ┃ ┃ ┃ ┣ 📜success.go
 ┃ ┃ ┃ ┗ 📜user.go
 ┃ ┃ ┃  📂utils
 ┃ ┃ ┃  ┣ 📜auth.go
 ┃ ┃ ┃ ┗ 📜consumer_wrapper.go
 ┃ ┣ 📂repositories
 ┃ ┃ ┣ 📜config.go
 ┃ ┃ ┗ 📜users_repository.go
 ┃ ┣ 📂resources
 ┃ ┣ 📂routes
 ┃ ┃ ┣ 📜auth_routes.go
 ┃ ┃ ┣ 📜diagnostics_routes.go
 ┃ ┃ ┣ 📜docs_routes.go
 ┃ ┃ ┣ 📜metrics_routes.go
 ┃ ┃ ┣ 📜swagger_routes.go
 ┃ ┃ ┗ 📜users_routes.go
 ┃ ┣ 📂services
 ┃ ┃ ┣ 📜auth_service.go
 ┃ ┃ ┣ 📜diagnostics_service.go
 ┃ ┃ ┣ 📜docs_service.go
 ┃ ┃ ┗ 📜users_service.go
 ┃ ┗ 📂utils
 ┃ ┃ ┣ 📜auth_utils.go
 ┃ ┃ ┗ 📜controller_utils.go
 ┣ 📂docker
 ┃ ┣ 📂certificates
 ┃ ┣ 📂docs
 ┃ ┃ ┣ 📜docs.go
 ┃ ┃ ┣ 📜swagger.json
 ┃ ┃ ┗ 📜swagger.yaml
 ┃ ┗ 📜Dockerfile
 ┣ 📂tests
 ┃ ┣ 📂registry
 ┃ ┃ ┣ 📜config.go
 ┃ ┃ ┣ 📜mocks.go
 ┃ ┃ ┣ 📜registry_function.go
 ┃ ┃ ┣ 📜registry_schema.go
 ┃ ┃ ┗ 📜registry_struct.go
 ┃ ┣ 📂resources
 ┃ ┃ ┣ 📂certificates
 ┃ ┃ ┣ 📂e2e
 ┃ ┃ ┣ 📂integration
 ┃ ┃ ┗ 📂unit
 ┃ ┃ ┃ ┣ 📜backups.json
 ┃ ┃ ┃ ┣ 📜config.json
 ┃ ┃ ┃ ┣ 📜controllers.json
 ┃ ┃ ┃ ┣ 📜infrastructure.json
 ┃ ┃ ┃ ┣ 📜middlewares.json
 ┃ ┃ ┃ ┣ 📜models.json
 ┃ ┃ ┃ ┣ 📜repositories.json
 ┃ ┃ ┃ ┣ 📜services.json
 ┃ ┃ ┃ ┗ 📜utils.json
 ┃ ┣ 📂settings
 ┃ ┃ ┣ 📜.mockery.yml
 ┃ ┃ ┗ 📜settings.yaml
 ┃ ┣ 📜main_test.go
 ┃ ┗ 📜runner_test.go
 ┣ 📜.gitignore
 ┣ 📜README.md
 ┣ 📜go.mod
 ┣ 📜go.sum
 ┗ 📜start.sh

```
## Testing
- **Running unit tests**
   ```bash
      go test -v -tags=unit ./tests -args -test_type=unit
   ```

- **Running integration tests**
   ```bash
      go test -v -tags=integration ./tests -args -test_type=integration
   ```

- **Running e2e tests**
   ```bash
      go test -v -tags=e2e ./tests -args -test_type=e2e
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


## License
- {TO DO}

## Notes
- {TO DO}
