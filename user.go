package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
		server: server,
	}
	// 启动一个协程来监听用户的消息
	go user.ListenMessage()
	return user
}

func (user *User) Online() {
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	user.server.Broadcast(user, "has joined the chat room \n")

}

func (user *User) Offline() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	user.server.Broadcast(user, "has left the chat room \n")
}

//给当前用户的客户端发送消息
func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		user.server.mapLock.Lock()
		for _, u := range user.server.OnlineMap {
			onlineMsg := "[" + u.Addr + "]" + u.Name + ": online...\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	}else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := msg[7:]
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("The name is already used, please change another name\n")
		}else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			user.SendMsg("Your name has been updated to: " + user.Name + "\n")
		}
	}else if len(msg) > 4 && msg[:3] == "to|"{
		//获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("Message format is incorrect, please use \"to|username|message\" format\n")
			return
		}
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("The user is not online\n")
			return
		}
		//获取消息内容
		content := strings.Split(msg, "|")[2]
		remoteUser.SendMsg(user.Name + " says: " + content + "\n")

	}else {
		user.server.Broadcast(user, msg)
	}
}



func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}
