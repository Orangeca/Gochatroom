package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	//连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("connect error", err)
		return nil
	}
	client.conn = conn
	//返回对象
	return client
}

func (client *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
	}
}

func (client *Client) PrivateChat() {
	client.SelectUser()
	fmt.Println(">>>>选择私聊目标用户")
	var remoteUsername string
	fmt.Scan(&remoteUsername)
	fmt.Println(">>>>输入聊天内容")
	var content string
	fmt.Scan(&content)
	sendMsg := "to|" + remoteUsername + "|" + content + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
	}
}

func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>>输入聊天内容，exit退出")
	fmt.Scan(&chatMsg)
	for chatMsg != "exit" {
		sendMsg := chatMsg + "\n"
		_, err := client.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("conn.Write err:", err)
			return
		}
		fmt.Println(">>>>输入聊天内容，exit退出")
		fmt.Scan(&chatMsg)
	}
}

func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn) //conn没有消息的时候就会阻塞
	//上面这句和下面5行等价
	//for{
	//	buf:=make([]byte,4096)
	//	client.conn.Read(buf)
	//	fmt.Println(string(buf))
	//}
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	fmt.Scan(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>输入新用户名")
	fmt.Scan(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}
		switch client.flag {
		case 1:
			//1.公聊模式
			fmt.Println("选择1.公聊模式")
			client.PublicChat()
		case 2:
			fmt.Println("选择2.私聊模式")
			client.PrivateChat()
		case 3:
			fmt.Println("选择3.修改username")
			client.UpdateName()
		}
	}
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置server IP")
	flag.IntVar(&serverPort, "port", 8888, "默认端口8888")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>连接失败...")
		return
	}
	fmt.Println(">>>>>连接成功...")
	//启动客户端的业务
	go client.DealResponse()
	client.Run()
}
