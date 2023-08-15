package main

import "C"
import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

// create a new server
func Newserver(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//将用户加入online map，并广播
	user := NewUser(conn, this)
	user.Online()
	//监听用户是否活跃
	isLive := make(chan bool)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := user.conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
				return
			}
			//提取用户信息去掉'\n'
			msg := string(buf[:n-1])
			//将得到的消息进行广播
			user.DoMessage(msg)
			isLive <- true
		}
	}()
	for {
		select {
		case <-isLive:
			//不做任何事情，为了激活select更新定时器
		case <-time.After(time.Second * 300):
			//超时强制关闭
			user.SendMsg("你被踢了")
			close(user.C)
			conn.Close()
			runtime.Goexit()
		}
	}
}

func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("ner.Listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()
	//listen message
	go this.ListenMessager()
	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener Accept err:", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}
}
