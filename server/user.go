package server

import (
	"fmt"
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

func NewUser(conn net.Conn, newServer *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: newServer,
	}
	go user.ListenMessage()
	return user
}

func (u *User) ListenMessage() {
	for {
		msg, ok := <-u.C
		if !ok {
			return
		}
		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("Write error: ", err)
			return
		}
	}
}

func (u *User) Online() {
	u.server.MapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.MapLock.Unlock()

	u.server.BroadCast(u, "已上线")
}

func (u *User) Offline() {
	u.server.MapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.MapLock.Unlock()

	u.server.BroadCast(u, "已下线")
}

func (u *User) DoMessage(msg string) {
	if msg == "who" {
		u.server.MapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMse := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			u.SendMessage(onlineMse)
		}
		u.server.MapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMessage("名字已被使用")
		} else {
			u.server.MapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.MapLock.Unlock()

			u.Name = newName
			u.SendMessage("您已经更新用户名：" + u.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		data := strings.Split(msg, "|")
		remoteName := data[1]
		if remoteName == "" {
			u.SendMessage("消息格式不正确")
			return
		}
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMessage("该用户不存在")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMessage("消息为空")
			return
		}
		remoteUser.SendMessage(u.Name + "对你说:" + content)
	} else {
		u.server.BroadCast(u, msg)
	}
}

func (u *User) SendMessage(msg string) {
	_, err := u.conn.Write([]byte(msg))
	if err != nil {
		return
	}
}
