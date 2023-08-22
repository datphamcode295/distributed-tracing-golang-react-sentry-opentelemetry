# Distributed Tracing Example
This is example how to implement distributed tracing in golang and reactjs with Opentelemetry and Sentry
## How to run
### Step 1: setup env
Copy .env from .env.example
### Step 2: run RabbitMQ docker
```
make docker-up
```
### Step 3: install all dependencies
```
make init
```
### Step 4: runing services
Open 3 terminals, then run:
```
make run-api-service
make run-mail-service   
make run-fe
```
