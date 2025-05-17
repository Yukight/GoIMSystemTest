package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播
	Message chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

func (server *Server) ListenMessage() {
	//监听广播消息
	for {
		msg := <-server.Message
		//遍历在线用户，发送消息
		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

func (server *Server) Broadcast(user *User, msg string) {
	//广播消息
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	// ...当前链接的业务处理
	// fmt.Printf("server connected to %s:%d, at: %s \n", server.Ip, server.Port, time.DateTime)

	//用户上线广播
	user := NewUser(conn, server)
	user.Online()

	//接收用户消息
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Printf("conn.Read err: %s", err)
				return
			}
			//提取用户消息
			msg := string(buf[:n-1])
			//广播消息
			user.DoMessage(msg)
		}
	}()

	select {}
}

func (server *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", server.Ip+":"+strconv.Itoa(server.Port))
	if err != nil {
		fmt.Printf("net.listen err: %s", err)
		return
	}
	// close listen socket(defer)
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Printf("listener.Close err: %s", err)
		}
	}(listener)
	go server.ListenMessage()
	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("listener.Accept err: %s", err)
			continue
		}

		// do handler
		go server.Handler(conn)
	}

}
