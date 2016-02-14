package main

import (
	"../utils"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

const PORT = "7777"

func main() {
	// 1 Connect
	conn, err := net.Dial("tcp", ":"+PORT)
	utils.CheckError(err, "Can't connect to server", true)
	client := NewClient(conn)
	defer conn.Close()
	//go
	watchForConnection(client)
}

///////// Client Class ////////////////////////////////////////////////////////

type Client struct {
	uid      string
	login    string
	outgoing chan []byte
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)
	client := &Client{
		uid:      "",
		login:    "",
		outgoing: make(chan []byte),
		reader:   reader,
		writer:   writer,
	}
	client.Listen()
	return client
}

// Listen - start corotinues for listening and writing
func (c *Client) Listen() {
	//go c.read()
	go c.write()
}

func (c *Client) write() {
	for data := range c.outgoing {
		c.writer.Write(data)
		c.writer.Flush()
	}
}

func (c *Client) Login() {
	s, err := json.Marshal(struct {
		Action string        `json:"action"`
		Data   utils.CltAuth `json:"data"`
	}{
		Action: "auth",
		Data:   utils.CltAuth{Login: "user", Pass: "12345"},
	})
	//fmt.Printf("Result %v\n", string(s))
	if utils.CheckError(err, "Login error", false) {
		c.writer.Write(s)
		c.writer.Flush()
	}
}

func (c *Client) Register() {
	m := utils.CltRegister{
		Nick: "владимир",
	}
	m.Login = "user"
	m.Pass = "12345"

	s, err := json.Marshal(struct {
		Action string            `json:"action"`
		Data   utils.CltRegister `json:"data"`
	}{
		Action: "register",
		Data:   m,
	})
	c.login = "user"
	c.uid = "user"
	if utils.CheckError(err, "Register error", false) {
		c.outgoing <- s
	}
}

func (c *Client) GetChannelList() {
}

func (c *Client) Enter() {
	m := utils.CltChannel{
		Channel: utils.GetMD5Hash("Public"),
	}
	//m.sid = ""
	//m.cid = ""

	s, err := json.Marshal(struct {
		Action string           `json:"action"`
		Data   utils.CltChannel `json:"data"`
	}{
		Action: "enter",
		Data:   m,
	})

	if utils.CheckError(err, "Register error", false) {
		c.outgoing <- s
	}
}

func (c *Client) Message() {
	m := utils.CltMessage{
		Body: "Текст на русском языке\nС переносом строк\n在中國文字",
	}
	m.Channel = utils.GetMD5Hash("Public")
	//m.sid = ""
	//m.cid = ""

	s, err := json.Marshal(struct {
		Action string           `json:"action"`
		Data   utils.CltMessage `json:"data"`
	}{
		Action: "message",
		Data:   m,
	})

	if utils.CheckError(err, "Register error", false) {
		c.outgoing <- s
	}
}

func watchForConnection(client *Client) {
	dec := json.NewDecoder(client.reader)
	count := 1
	for {
		var m utils.SrvMessage
		err := dec.Decode(&m)
		utils.CheckError(err, "Invalid message", true)
		switch m.Action {
		case "welcome":
			fmt.Println("Welcome message")
			client.Login()
			//client.Register()
		case "auth": //, "register", "leave":
			fmt.Printf("Message %v\n", m.Action)
			var im utils.SrvStatusMessage
			if err := json.Unmarshal(m.RawData, &im); err != nil {
				log.Fatal(err)
				return
			}
			if im.Status == 0 {
				client.Enter()
			} else {
				fmt.Printf("Auth error - %v", im.Error)
				client.Login()
			}
		case "enter":
			fmt.Printf("enter count - %v\n", count)
			client.Message()
		case "ev_enter":
			fmt.Printf("ev_enter count - %v\n", count)
		case "ev_message":
			if count < 10 {
				time.Sleep(time.Second)
				fmt.Printf("ev_message count - %v\n", count)
				client.Message()
				count += 1
			}
		case "channellist":
			fmt.Printf("channellist Message\n")
		default:
			fmt.Printf("Unknown Message %v\n", m.Action)
		}
	}
}
