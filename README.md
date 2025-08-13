# Ingress App
This is a simple web application built with the Go language and the Echo framework. The app acts as a producer, receiving file with HTTP request and uploading it to MinIO and publishing a message to a Kafka topic.

## 🚀 Getting Started
1. Start the Cluster
This project uses a single-node Kafka cluster running in KRaft mode (without ZooKeeper) and a web-based UI for management. It also includes a MinIO server for S3-compatible object storage. To start the cluster, navigate to the project's root directory in your terminal and run:
  ```sh
  docker compose up -d
  ```
This command will start the Kafka broker, the Kafka UI, and the MinIO server. The Kafka UI will be available at http://localhost:8080. The MinIO Console will be available at http://localhost:9001.

2. Install Go Dependencies
Next, you need to install the Go packages required by the application. The go mod tidy command reads the go.mod file and downloads all the necessary dependencies.
```sh
go mod tidy
```
3. Run the Server
Once the dependencies are installed, you can start the Go server.
```sh
go run .
```
The server will accept requests at http://localhost:5000.

## ⚙️ How It Works
The application has a single HTTP endpoint `/publish` that upload the file to MinIO and produce a message to Kafka.

### The Endpoint
* Method: POST
* URL: http://localhost:5000/publish
* Body: A form-data payload that accepts a `file` as a key and the file itself as the value.

After running this command, the Go application will upload the file to MinIO and produce a message to the files Kafka topic with the filename as the key. You can then use the Kafka UI at http://localhost:8080 to view the message in the topic.

### MinIO Server
The docker-compose.yml file also starts a MinIO server.

* Endpoint: The S3 API is available at http://localhost:9000.

* Console: The MinIO web console is available at http://localhost:9001.

* Credentials:
  *  root access: `minioadmin`,
  *  root secret: `minioadmin`.