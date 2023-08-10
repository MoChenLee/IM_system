package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	MapLock   sync.RWMutex

	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,

		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (s *Server) Handler(conn net.Conn) {
	fmt.Println("建立连接成功")
	newUser := NewUser(conn, s)
	newUser.Online()

	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			if n == 0 {
				newUser.Offline()
				return
			}
			msg := string(buf[:n-1])
			newUser.DoMessage(msg)
			isLive <- true
		}
	}()
	//s.BroadCast(newUser, "已上线")
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 10):
			newUser.SendMessage("长时间无操作,结束")
			close(newUser.C)
			s.MapLock.Lock()
			delete(s.OnlineMap, newUser.Name)
			s.MapLock.Unlock()
			err := conn.Close()
			if err != nil {
				return
			}
			return
		}
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	defer func(listener net.Listener) {
		err = listener.Close()
		if err != nil {
			fmt.Println("net.Listener Close err:", err)
		}
	}(listener)

	go s.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("net.Listener Accept err:", err)
			continue
		}
		go s.Handler(conn)
	}

}

func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message
		s.MapLock.Lock()
		for _, m := range s.OnlineMap {
			m.C <- msg
		}
		s.MapLock.Unlock()
	}
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}
