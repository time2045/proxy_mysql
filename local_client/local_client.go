package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"github.com/gorilla/websocket"
)

// Config 结构体用于映射 JSON 配置
type ClientConfig struct {
	LocalListenAddress string `json:"local_listen_address"`
	ServerWsUrl        string `json:"server_ws_url"`
	LogPath            string `json:"log_path"`
}

// 全局配置变量
var config ClientConfig

func handleLocalConnection(localConn net.Conn) {
	defer localConn.Close()
	log.Printf("接收到新的本地连接来自: %s", localConn.RemoteAddr())

	wsConn, _, err := websocket.DefaultDialer.Dial(config.ServerWsUrl, nil)
	if err != nil {
		log.Printf("连接WebSocket服务器 %s 失败: %v", config.ServerWsUrl, err)
		return
	}
	defer wsConn.Close()
	log.Println("已通过 WebSocket 成功连接到服务器")

	errChan := make(chan error, 2)

	// Local TCP -> WebSocket
	go func() {
		_, err := io.Copy(wsMessageWriter{wsConn}, localConn)
		if err != nil {
			log.Printf("从本地连接读取数据并写入 WebSocket 时出错: %v", err)
		}
		wsConn.Close()
		errChan <- err
	}()

	// WebSocket -> Local TCP
	go func() {
		for {
			mt, message, err := wsConn.ReadMessage()
			if err != nil {
				log.Printf("从 WebSocket 读取数据时出错: %v", err)
				localConn.Close()
				errChan <- err
				return
			}
			if mt == websocket.BinaryMessage {
				if _, err := localConn.Write(message); err != nil {
					log.Printf("写入本地连接时出错: %v", err)
					errChan <- err
					return
				}
			}
		}
	}()

	<-errChan
	log.Println("本地连接已关闭:", localConn.RemoteAddr())
}

// wsMessageWriter 结构体实现了 io.Writer 接口
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
	configFile, err := os.Open("config.client.json")
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
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	listener, err := net.Listen("tcp", config.LocalListenAddress)
	if err != nil {
		log.Fatalf("监听 %s 失败: %v", config.LocalListenAddress, err)
	}
	defer listener.Close()

	log.Println("------------------------------------")
	log.Println("本地 TCP 到 WebSocket 客户端已启动。")
	log.Printf("配置已加载：正在监听 %s，连接到 %s", config.LocalListenAddress, config.ServerWsUrl)
	log.Println("请配置您的 MySQL 客户端 (Navicat) 连接到:", config.LocalListenAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}
		go handleLocalConnection(conn)
	}
}