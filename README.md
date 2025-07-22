# RESTful OTP Authentication API

A secure, robust, and production-ready RESTful API built with Go for user authentication using phone numbers and one-time passwords (OTP). This project demonstrates a clean, layered architecture, containerization with Docker, and a full suite of security features.

## Features

-   **Secure Sign-up & Sign-in:** Users can register and log in using only their phone number.
-   **OTP Verification:** A time-sensitive, 6-digit OTP is used to verify phone number ownership.
-   **JWT-based Authentication:** Upon successful verification, a JSON Web Token (JWT) is issued for authenticating subsequent API requests.
-   **Robust Security:**
    -   **Rate Limiting:** Prevents SMS spam by limiting how often a user can request an OTP.
    -   **Attempt Limiting:** Prevents online brute-force attacks by invalidating an OTP after 10 failed verification attempts.
    -   **Secure Password Hashing:** Uses the `bcrypt` algorithm to securely hash and store OTPs.
    -   **Time-Limited OTPs:** Each OTP is valid for only 5 minutes.
-   **Containerized:** Fully containerized using Docker and Docker Compose for easy setup and deployment.
-   **Comprehensive API Documentation:** Includes a `swagger.yaml` (OpenAPI 3.0) file for clear, interactive API documentation.
-   **Unit Tested:** The core business logic is thoroughly unit-tested using mocks to ensure reliability.

## Architecture

This project follows a clean, layered architecture to ensure a strong separation of concerns, making the codebase maintainable, scalable, and testable.

-   `main.go`: The entry point of the application. Responsible for reading configuration, setting up the database connection, and performing dependency injection to wire all the layers together.
-   `internal/handlers`: The presentation layer. Responsible for handling HTTP requests, decoding request bodies, validating input, and returning JSON responses. It knows nothing about the business logic, only how to talk HTTP.
-   `internal/services`: The business logic layer. Contains all the core application logic (e.g., checking rate limits, verifying OTPs, creating users, generating JWTs). It knows nothing about HTTP or the database implementation.
-   `internal/repository`: The data access layer. Responsible for all communication with the PostgreSQL database. It translates requests from the service layer into SQL queries.
-   `internal/models`: Defines the Go structs that represent our database entities (`User`, `OTP`).

## API Documentation

The API is documented using the OpenAPI 3.0 standard in the `api/swagger.yaml` file.

You can use a tool like the [Swagger Editor](https://editor.swagger.io/) to view the documentation, or you can import the `swagger.yaml` file directly into Postman to create a pre-configured collection for testing.

### Endpoints

-   `POST /api/v1/auth/initiate`: Initiates the login/sign-up process and sends an OTP.
-   `POST /api/v1/auth/verify`: Verifies the OTP and returns a JWT.

## Getting Started

### Prerequisites

-   [Docker](https://www.docker.com/get-started)

### Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/ari-ahm/restful-otp
    cd restful-otp
    ```

2.  **Run the application:**
    Use Docker Compose to build the Go application image and start the API and PostgreSQL containers.

    ```bash
    docker compose up --build
    ```

    The API will be running and accessible at `http://localhost:8080`.

3.  **Test with Postman:**
    -   Import the `api/swagger.yaml` file into Postman.
    -   Follow the two-step process:
        1.  Send a `POST` request to the `Initiate` endpoint with a phone number.
        2.  Check the terminal logs for the simulated OTP.
        3.  Send a `POST` request to the `Verify` endpoint with the phone number and the OTP from the logs to receive your JWT.

### Cleaning Up

To stop the containers and remove the database volume (for a complete reset), run:

```bash
docker compose down -v
```

## Testing

The project includes unit tests for the core service layer. To run the tests, execute the following command from the project root:

```bash
go test ./...
```

This command will discover and run all `_test.go` files in the project.

## Environment Variables

The application is configured using environment variables, which are set in the `compose.yml` file for local development.

| Variable                 | Description                                                  | Default Value (in `compose.yml`) |
| ------------------------ | ------------------------------------------------------------ | --------------------------------------- |
| `PORT`                   | The port on which the Go API server will listen.             | `8080`                                  |
| `JWT_SECRET_KEY`         | A long, secret key used for signing JWTs.                    | `a-very-secure-secret-key...`           |
| `DB_HOST`                | The hostname of the database server.                         | `db` (the service name in Docker)       |
| `DB_PORT`                | The port of the database server.                             | `5432`                                  |
| `DB_USER`                | The username for the database connection.                    | `postgres`                              |
| `DB_PASSWORD`            | The password for the database connection.                    | `mysecretpassword`                      |
| `DB_NAME`                | The name of the database to use.                             | `restful_otp_db`                        |


## Technology Choices & Trade-offs

### Why PostgreSQL?

PostgreSQL was chosen for its optimal balance of reliability and simplicity for this application's needs.

While an in-memory store like **Redis** would be technically faster, the primary performance bottleneck in an OTP system is always the external SMS network call, which can take several seconds. The 1-2 milliseconds saved by using Redis would be imperceptible to the end-user.

In exchange for this negligible performance gain, using Redis would introduce significant architectural complexity, requiring the management of two separate databases.

Therefore, PostgreSQL is the more robust and maintainable choice, providing world-class data integrity (ACID compliance) with more than sufficient performance for this workload.
