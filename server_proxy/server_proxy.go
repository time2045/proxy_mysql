package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"github.com/gorilla/websocket"
)

// Config 结构体用于映射 JSON 配置
type ServerConfig struct {
	ListenAddress      string `json:"listen_address"`
	MysqlServerAddress string `json:"mysql_server_address"`
	LogPath            string `json:"log_path"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 全局变量，方便在 handler 中使用
var config ServerConfig

func handleConnection(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("升级到 WebSocket 失败: %v", err)
		return
	}
	defer wsConn.Close()
	log.Println("新的 WebSocket 客户端已连接，来自:", r.RemoteAddr)

	mysqlConn, err := net.Dial("tcp", config.MysqlServerAddress)
	if err != nil {
		log.Printf("连接 MySQL 服务器 %s 失败: %v", config.MysqlServerAddress, err)
		return
	}
	defer mysqlConn.Close()
	log.Println("已成功连接到 MySQL 服务器")

	errChan := make(chan error, 2)

	// WebSocket -> TCP
	go func() {
		for {
			mt, message, err := wsConn.ReadMessage()
			if err != nil {
				log.Printf("从 WebSocket 读取数据时出错: %v", err)
				mysqlConn.Close()
				errChan <- err
				return
			}
			if mt == websocket.BinaryMessage {
									if _, err := mysqlConn.Write(message); err != nil {
						log.Printf("写入 MySQL 时出错: %v", err)
					errChan <- err
					return
				}
			}
		}
	}()

	// TCP -> WebSocket
	go func() {
        // 使用 io.Copy 简化代码，并提高效率
		_, err := io.Copy(wsMessageWriter{wsConn}, mysqlConn)
		if err != nil {
			log.Printf("从 MySQL 读取数据并写入 WebSocket 时出错: %v", err)
		}
		wsConn.Close()
		errChan <- err
	}()
	
	// 等待任一方向的goroutine结束
	<-errChan
	log.Println("连接已关闭，来自:", r.RemoteAddr)
}

// wsMessageWriter 结构体实现了 io.Writer 接口，以便 io.Copy 可以使用
type wsMessageWriter struct {
	*websocket.Conn
}

func (w wsMessageWriter) Write(p []byte) (int, error) {
	err := w.Conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}


func main() {
	// 读取配置文件
	configFile, err := os.Open("config.server.json")
	if err != nil {
		log.Fatalf("打开配置文件失败: %v", err)
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	// 设置日志输出
	logFile, err := os.OpenFile(config.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("打开日志文件失败: %v", err)
	}
	// 将日志同时输出到文件和控制台
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	log.Println("------------------------------------")
	log.Println("WebSocket 到 TCP 代理服务器正在启动...")
	log.Printf("配置已加载：正在监听 %s，代理到 %s", config.ListenAddress, config.MysqlServerAddress)

	http.HandleFunc("/mysql", handleConnection)
	if err := http.ListenAndServe(config.ListenAddress, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}