# MySQL WebSocket Proxy

[中文](README_CN.md)

[![Go](https://img.shields.io/badge/Go-1.24.1+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

A simple yet powerful MySQL proxy tool written in Go that tunnels MySQL protocol traffic over WebSocket. This is particularly useful for bypassing network restrictions that limit connections to standard HTTP/WebSocket ports.

---

## ✨ Features

- **Client-Server Architecture**: Consists of a `local_client` and a `server_proxy`.
- **TCP to WebSocket Tunneling**: The client converts local TCP connections into WebSocket messages.
- **WebSocket to TCP Forwarding**: The server converts WebSocket messages back to TCP and forwards them to the target MySQL server.
- **Configuration Driven**: Easy to configure using JSON files.
- **Logging**: Both client and server log their activities to files.

---

## 📂 Project Structure

```
proxy_mysql/
├── go.mod
├── go.sum
├── README.md
├── README_CN.md
├── local_client/
│   ├── config.client.json
│   └── local_client.go
└── server_proxy/
    ├── config.server.json
    └── server_proxy.go
```

---

## 🚀 Getting Started

### Prerequisites

- Go 1.24.1+
- `gorilla/websocket` library

Before you begin, ensure you have Go installed. Then, download the required dependencies:

```bash
go mod tidy
```

### Configuration

You need to configure both the client and the server by editing their respective JSON configuration files.

#### Client Configuration (`local_client/config.client.json`)

```json
{
  "local_listen_address": "127.0.0.1:3307",
  "server_ws_url": "ws://YOUR_SERVER_IP:16781/mysql",
  "log_path": "local_client.log"
}
```

- `local_listen_address`: The local TCP address the client listens on. Your MySQL client (e.g., Navicat, DBeaver) will connect to this address.
- `server_ws_url`: The WebSocket URL of your remote `server_proxy`.
- `log_path`: Path to the client's log file.

#### Server Configuration (`server_proxy/config.server.json`)

```json
{
  "listen_address": "0.0.0.0:9090",
  "mysql_server_address": "127.0.0.1:3306",
  "log_path": "server_proxy.log"
}
```

- `listen_address`: The address the `server_proxy` listens on for incoming WebSocket connections.
- `mysql_server_address`: The address of your actual MySQL server.
- `log_path`: Path to the server's log file.

### Compilation

You can compile the client and server manually for your current operating system, cross-compile for other platforms, or use the provided PowerShell script to build for all supported platforms at once.

#### Using the Build Script (Recommended)

On Windows, you can use the provided PowerShell script to compile the client and server for all target platforms (Windows, Linux, macOS).

```powershell
.\build.ps1
```

After running the script, you will find all the compiled binaries in the `builds` directory.

#### Manual Compilation

If you prefer to compile manually, follow these instructions.

##### Compile for Current OS

- **Build `server_proxy`**:
  ```bash
  go build -o server_proxy ./server_proxy/
  ```
- **Build `local_client`**:
  ```bash
  go build -o local_client ./local_client/
  ```

##### Cross-Compile for Linux (amd64)

- **Build `server_proxy` for Linux**:
  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server_proxy_linux ./server_proxy/
  ```
- **Build `local_client` for Linux**:
  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o local_client_linux ./local_client/
  ```

### Running the Proxy

1.  **On your server**: Run the `server_proxy`.
    ```bash
    ./server_proxy
    ```

2.  **On your local machine**: Run the `local_client`.
    ```bash
    ./local_client
    ```

3.  **Connect your MySQL client**: Configure your MySQL client to connect to the `local_listen_address` specified in `config.client.json` (e.g., `127.0.0.1:3307`).

---

## 🔌 Nginx Configuration (Optional)

If you want to run the `server_proxy` behind Nginx (e.g., for SSL termination or to share port 80/443), you can use the following configuration:

```nginx
server {
    listen 16781; # Or your desired public port (e.g., 80, 443)
    server_name your_domain.com;

    location /mysql {
        # Forward requests to the server_proxy
        proxy_pass http://127.0.0.1:9090; # Must match listen_address in config.server.json

        # Required for WebSocket
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;

        # Increase timeouts for long-lived connections
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
```

---

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](https://opensource.org/licenses/MIT) file for details.
