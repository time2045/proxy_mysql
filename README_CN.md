# MySQL WebSocket 代理

[English](README.md)

[![Go](https://img.shields.io/badge/Go-1.24.1+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

一个使用 Go 语言编写的简单而强大的 MySQL 代理工具，它通过 WebSocket 隧道传输 MySQL 协议流量。这在需要绕过仅限标准 HTTP/WebSocket 端口的网络限制时特别有用。

---

## ✨ 功能特性

- **客户端-服务器架构**: 包含一个 `local_client`（本地客户端）和一个 `server_proxy`（服务器代理）。
- **TCP -> WebSocket 隧道**: 客户端将本地 TCP 连接转换为 WebSocket 消息。
- **WebSocket -> TCP 转发**: 服务器将 WebSocket 消息转回 TCP 流量，并转发到目标 MySQL 服务器。
- **JSON 文件配置**: 使用 JSON 文件轻松配置。
- **日志记录**: 客户端和服务器都会将运行日志记录到文件中。

---

## 📂 项目结构

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

## 🚀 快速开始

### 先决条件

- Go 1.24.1 或更高版本
- `gorilla/websocket` 库

在开始之前，请确保您已安装 Go。然后，下载所需的依赖项：

```bash
go mod tidy
```

### 配置

您需要通过编辑相应的 JSON 配置文件来配置客户端和服务器。

#### 客户端配置 (`local_client/config.client.json`)

```json
{
  "local_listen_address": "127.0.0.1:3307",
  "server_ws_url": "ws://你的服务器IP:16781/mysql",
  "log_path": "local_client.log"
}
```

- `local_listen_address`: 客户端监听的本地 TCP 地址。您的 MySQL 客户端（如 Navicat、DBeaver）将连接到此地址。
- `server_ws_url`: 远程 `server_proxy` 的 WebSocket URL。
- `log_path`: 客户端日志文件的路径。

#### 服务器配置 (`server_proxy/config.server.json`)

```json
{
  "listen_address": "0.0.0.0:9090",
  "mysql_server_address": "127.0.0.1:3306",
  "log_path": "server_proxy.log"
}
```

- `listen_address`: `server_proxy` 用于监听传入 WebSocket 连接的地址。
- `mysql_server_address`: 你的实际 MySQL 服务器地址。
- `log_path`: 服务器日志文件的路径。

### 编译

您可以手动为当前操作系统编译客户端和服务器，为其他平台进行交叉编译，或者使用提供的 PowerShell 脚本一次性为所有支持的平台进行构建。

#### 使用构建脚本 (推荐)

在 Windows 上，您可以使用提供的 PowerShell 脚本为所有目标平台（Windows、Linux、macOS）编译客户端和服务器。

```powershell
.\build.ps1
```

运行脚本后，您将在 `builds` 目录中找到所有已编译的二进制文件。

#### 手动编译

如果您喜欢手动编译，请按照以下说明操作。

##### 为当前操作系统编译
=======
您可以为当前操作系统编译客户端和服务器，或为其他平台（如 Linux）进行交叉编译。

#### 为当前操作系统编译

- **编译 `server_proxy`**:
  ```bash
  go build -o server_proxy ./server_proxy/
  ```
- **编译 `local_client`**:
  ```bash
  go build -o local_client ./local_client/
  ```

##### 交叉编译 Linux (amd64) 版本
=======
#### 交叉编译 Linux (amd64) 版本

- **为 Linux 编译 `server_proxy`**:
  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server_proxy_linux ./server_proxy/
  ```
- **为 Linux 编译 `local_client`**:
  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o local_client_linux ./local_client/
  ```

### 运行代理

1.  **在您的服务器上**: 运行 `server_proxy`。
    ```bash
    ./server_proxy
    ```

2.  **在您的本地计算机上**: 运行 `local_client`。
    ```bash
    ./local_client
    ```

3.  **连接您的 MySQL 客户端**: 配置您的 MySQL 客户端连接到 `config.client.json` 中指定的 `local_listen_address`（例如 `127.0.0.1:3307`）。

---

## 🔌 Nginx 配置 (可选)

如果您希望在 Nginx 后面运行 `server_proxy`（例如，为了 SSL 终止或共享 80/443 端口），您可以使用以下配置：

```nginx
server {
    listen 16781; # 或者您期望的公共端口 (例如 80, 443)
    server_name your_domain.com;

    location /mysql {
        # 将请求转发到 server_proxy
        proxy_pass http://127.0.0.1:9090; # 必须与 config.server.json 中的 listen_address 匹配

        # WebSocket 所需的头信息
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;

        # 为长连接增加超时时间
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
```

---

## 📄 许可证

该项目根据 MIT 许可证授权。有关详细信息，请参阅 [LICENSE](https://opensource.org/licenses/MIT) 文件。
