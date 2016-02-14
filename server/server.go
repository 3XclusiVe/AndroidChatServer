package main

import (
	"../utils"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// PORT of Server
const PORT = "7777"

// Error codes
const (
	ErrOK              = 0  // All OK
	ErrAlreadyExist    = 1  // Login or Nickname or Channel already exist
	ErrInvalidPass     = 2  // Invalid login or password
	ErrInvalidData     = 3  // Invalid JSON
	ErrEmptyField      = 4  // Empty Nick, Login, Password or Channel
	ErrAlreadyRegister = 5  // User is already registered
	ErrNeedAuth        = 6  // User has to auth
	ErrNeedRegister    = 7  // User has to register
	ErrUserNotFound    = 8  // User not found by uid
	ErrChannelNotFound = 9  // Channel not found by id
	ErrInvalidChannel  = 10 // User isn't in the channel
)

var server = NewServer()

// Main function
func main() {
	psock, err := net.Listen("tcp", ":"+PORT)
	utils.CheckError(err, "Can't create a server", true)
	fmt.Printf("Server start on port %v \n", PORT)
	for {
		conn, err := psock.Accept()
		if utils.CheckError(err, "Can't create connection", false) {
			client := newClient(conn)
			_ = client
		}
	}
}

////////////// Client Class ///////////////////////////////////////////////////

// Client class of client
type Client struct {
	conn     net.Conn
	uid      string
	login    string
	ip       string
	cid      string
	sid      string
	nick     string
	status   string
	outgoing chan []byte
	reader   *bufio.Reader
	writer   *bufio.Writer
	channels map[string]string
}

// write - send data to user
func (c *Client) write() {
	for data := range c.outgoing {
		c.writer.Write(data)
		c.writer.Flush()
	}
}

// Listen - start corotinues for listening and writing
func (c *Client) Listen() {
	go c.read()
	go c.write()
}

// newClient create new instance of Client class
func newClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)
	client := &Client{
		conn:     connection,
		uid:      "",
		login:    "",
		status:   "",
		ip:       connection.RemoteAddr().String(),
		outgoing: make(chan []byte),
		reader:   reader,
		writer:   writer,
		channels: make(map[string]string),
	}
	client.Listen()
	return client
}

// Register register new user on server
func (c *Client) Register(login string, pass string, nick string) {
	status, err := server.Register(c, login, pass, nick)
	if err != nil {
		c.Error("register", err.Error(), status, true)
		return
	}
	s, err := json.Marshal(struct {
		Action string                 `json:"action"`
		Data   utils.SrvStatusMessage `json:"data"`
	}{
		Action: "register",
		Data: utils.SrvStatusMessage{
			Status: ErrOK,
			Error:  "OK",
		},
	})
	if !c.CheckError(err, "Can't marhsal answer") {
		return
	}
	c.outgoing <- s

	c.Auth(login, pass)
}

// Disconnect is function a wrapper of conn.Close
func (c *Client) Disconnect() {
	c.conn.Close()
	server.LeaveChannel(c, "*")
	c.removeChannel("*")
}

// CheckError wrapper to check err construction
func (c *Client) CheckError(err error, message string) bool {
	if err != nil {
		fmt.Fprintln(os.Stderr, "("+c.ip+") "+message+": "+err.Error())
	}
	return err == nil
}

// Auth client autorisation on server
func (c *Client) Auth(login string, pass string) bool {
	sid, status, err := server.Auth(c, login, pass)
	if err != nil {
		c.Error("auth", err.Error(), status, true)
		return false
	}
	c.login = login
	c.uid = login
	c.sid = sid
	m := utils.SrvStatusAuthMessage{
		Sid:  c.sid,
		Cid:  c.cid,
		Nick: c.nick,
	}
	m.Status = ErrOK
	m.Error = "OK"
	s, err := json.Marshal(struct {
		Action string                     `json:"action"`
		Data   utils.SrvStatusAuthMessage `json:"data"`
	}{
		Action: "auth",
		Data:   m,
	})
	if !c.CheckError(err, "Can't marhsal message") {
		c.Disconnect()
		return false
	}
	c.outgoing <- s
	return true
}

// Ok sends to client a status OK
func (c *Client) Ok(action string) {
	s, err := json.Marshal(struct {
		Action string                 `json:"action"`
		Data   utils.SrvStatusMessage `json:"data"`
	}{
		Action: action,
		Data: utils.SrvStatusMessage{
			Status: ErrOK,
			Error:  "OK",
		},
	})
	if !c.CheckError(err, "Can't marhsal message") {
		c.Disconnect()
		return
	}
	c.outgoing <- s
}

// Error sends to client an error status
func (c *Client) Error(action string, text string, status int, closeConn bool) {
	message := utils.SrvStatusMessage{
		Status: status,
		Error:  text,
	}
	data, err := json.Marshal(struct {
		Action string                 `json:"action"`
		Data   utils.SrvStatusMessage `json:"data"`
	}{
		Action: action, Data: message,
	})
	if !c.CheckError(err, "Can't marhsal message") {
		c.Disconnect()
		return
	}
	fmt.Printf("Error: from %v - %v\n", c.ip, text)
	c.writer.Write(data)
	c.writer.Flush()
	if closeConn {
		c.Disconnect()
	}
}

func (c *Client) read() {
	message := utils.SrvWelcomeMessage{
		Action:  "welcome",
		Time:    int(time.Now().Unix()),
		Message: "Welcome to chat\n server",
	}
	start, err := json.Marshal(message)
	if !c.CheckError(err, "Can't marhsal message") {
		c.Error("unknown", "Invalid request", ErrInvalidData, true)
		return
	}
	c.outgoing <- start
	fmt.Printf("Send message to client\n")
	dec := json.NewDecoder(c.reader)
	for {
		var m utils.CltRequest
		err := dec.Decode(&m)
		if !c.CheckError(err, "Invalid message\n") {
			c.Error("unknown", "Invalid request", ErrInvalidData, true)
			return
		}
		fmt.Printf("Action %v, %v\n", m.Action, string(m.RawData))
		if c.uid == "" && (m.Action != "register" && m.Action != "auth") {
			c.Error(m.Action, "Need auth", ErrNeedAuth, false)
			continue
		}
		switch m.Action {
		case "register":
			var im utils.CltRegister
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Register: Invalid data", ErrInvalidData, true)
				return
			}
			if c.login != "" {
				c.Error(m.Action, "Already register", ErrAlreadyRegister, true)
				return
			}
			c.Register(im.Login, im.Pass, im.Nick)
		case "auth":
			var im utils.CltAuth
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Auth: Invalid data", ErrInvalidData, true)
				return
			}
			if !c.Auth(im.Login, im.Pass) {
				return
			}
		case "setuserinfo":
			var im utils.CltSetUserInfo
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "SetUserInfo: Invalid data", ErrInvalidData, true)
				return
			}
			c.status = im.UserStatus
			c.Ok(m.Action)

		case "channellist":
			var im utils.CltBaseReq
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Channels Invalid data", ErrInvalidData, true)
				return
			}
			server.GetChannelList(c)
		case "createchannel":
			var im utils.CltCreateChannel
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "CreateChannels Invalid data", ErrInvalidData, true)
				return
			}
			server.CreateChannel(c, im.Name, im.Descr)
		case "userinfo":
			var im utils.CltUserInfo
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Invalid data", ErrInvalidData, true)
				return
			}
			server.GetUserInfo(c, im.User)
		case "enter":
			var im utils.CltChannel
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Invalid data", ErrInvalidData, true)
				return
			}
			server.EnterToChannel(c, im.Channel)
		case "leave":
			var im utils.CltChannel
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Invalid data", ErrInvalidData, true)
				return
			}
			server.LeaveChannel(c, im.Channel)
		case "message":
			var im utils.CltMessage
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Invalid data", ErrInvalidData, true)
				return
			}
			server.SendMessage(c, im.Channel, im.Body)
		default:
		}
	}
}

// addChannel add channel info to user
func (c *Client) addChannel(chid string) {
	c.channels[chid] = chid
}

// removeChannel revmove channel info from user
func (c *Client) removeChannel(chid string) {
	if chid == "*" {
		c.channels = make(map[string]string)
	} else {
		delete(c.channels, chid)
	}
}

///////////////// Server Class ////////////////////////////////////////////////

// Server is global data storage
type Server struct {
	Logins       map[string]string
	Nicks        map[string]string
	LoginsPasses map[string]string
	Users        map[string]string
	Clients      map[string]*Client
	Channels     map[string]*Channel
}

// NewServer is constructor of Server
func NewServer() *Server {
	s := &Server{
		Logins:       make(map[string]string),
		Nicks:        make(map[string]string),
		LoginsPasses: make(map[string]string),
		Users:        make(map[string]string),
		Clients:      make(map[string]*Client),
		Channels:     make(map[string]*Channel),
	}

	ch := NewChannel("Public", "Public")
	s.Channels[ch.id] = ch

	return s
}

// Auth checks auth of Client
func (s *Server) Auth(c *Client, login string, pass string) (string, int, error) {

	nick, ok := s.Logins[login]
	if !ok {
		return "", ErrNeedRegister, errors.New("Need to register")
	}
	p, ok := s.LoginsPasses[login]
	if !ok || p != pass {
		return "", ErrInvalidPass, errors.New("Invalid login or password!")
	}
	s.Clients[login] = c
	c.nick = nick
	c.cid = login

	return utils.GetMD5Hash(login), ErrOK, nil
}

// Register adds new user
func (s *Server) Register(c *Client, login string, pass string, nick string) (int, error) {
	if login == "" || nick == "" || pass == "" {
		return ErrEmptyField, errors.New("Empty field")
	}
	_, ok := s.Nicks[nick]
	if ok {
		return ErrAlreadyExist, errors.New("Nick already was used")
	}
	_, ok = s.Logins[login]
	if ok {
		return ErrAlreadyExist, errors.New("Login already was used")
	}
	s.Nicks[nick] = login
	s.Logins[login] = nick
	s.LoginsPasses[login] = pass
	c.cid = login
	s.Users[c.cid] = login
	return ErrOK, nil
}

// GetChannelList send to user list of channel
func (s *Server) GetChannelList(c *Client) {
	list := utils.SrvListOfChannel{}
	list.Status = ErrOK
	list.Error = "OK"
	list.Channels = make([]utils.ChannelData, 0)

	for id, ch := range s.Channels {
		list.Channels = append(list.Channels, utils.ChannelData{
			Id:     id,
			Name:   ch.name,
			Descr:  ch.descr,
			Online: len(ch.clients),
		})
	}

	m, err := json.Marshal(struct {
		Action string                 `json:"action"`
		Data   utils.SrvListOfChannel `json:"data"`
	}{
		Action: "channellist",
		Data:   list,
	})

	if !c.CheckError(err, "Can't marhsal answer") {
		return
	}
	c.outgoing <- m
}

// CreateChannel trying to create new channel
func (s *Server) CreateChannel(c *Client, name string, descr string) {
	id := utils.GetMD5Hash(name)
	if name == "" {
		c.Error("createchannel", "Name of channel is empty", ErrEmptyField, false)
		return
	}
	if _, ok := s.Channels[id]; ok {
		c.Error("createchannel", "Channel already exist", ErrAlreadyExist, false)
		return
	}
	ch := NewChannel(name, descr)
	s.Channels[ch.id] = ch

	m := utils.SrvAddChannelMessage{ChannelID: ch.id}
	m.Status = ErrOK
	m.Error = "OK"

	sm, err := json.Marshal(struct {
		Action string                     `json:"action"`
		Data   utils.SrvAddChannelMessage `json:"data"`
	}{
		Action: "createchannel",
		Data:   m,
	})
	if !c.CheckError(err, "Can't marhsal message") {
		c.Disconnect()
		return
	}
	c.outgoing <- sm
}

// GetUserInfo gets user info to another user
func (s *Server) GetUserInfo(c *Client, uid string) {
	client, ok := s.Clients[uid]
	if !ok {
		c.Error("userinfo", "User not found", ErrUserNotFound, false)
		return
	}
	m := utils.SrvUserInfo{
		Nick:       client.nick,
		UserStatus: client.status,
	}
	m.Status = ErrOK
	m.Error = "OK"

	mess, err := json.Marshal(struct {
		Action string            `json:"action"`
		Data   utils.SrvUserInfo `json:"data"`
	}{
		Action: "userinfo",
		Data:   m,
	})

	if !c.CheckError(err, "Can't marhsal message") {
		c.Disconnect()
		return
	}
	c.outgoing <- mess
}

// EnterToChannel user enters to a channel
func (s *Server) EnterToChannel(c *Client, chid string) {
	ch, ok := s.Channels[chid]
	if !ok {
		c.Error("enter", "Channel not found", ErrChannelNotFound, false)
		return
	}

	ans := utils.SrvEnterToChannel{}
	ans.Status = ErrOK
	ans.Error = "OK"
	ans.Users = make([]utils.UserData, 0)
	ans.LastMsgs = make([]utils.MessageData, 0)

	for _, client := range ch.clients {
		ans.Users = append(ans.Users, utils.UserData{
			Uid:  client.cid,
			Nick: client.nick,
		})
	}

	m, err := json.Marshal(struct {
		Action string                  `json:"action"`
		Data   utils.SrvEnterToChannel `json:"data"`
	}{
		Action: "enter",
		Data:   ans,
	})
	if !c.CheckError(err, "Can't marhsal answer") {
		return
	}
	c.outgoing <- m

	ch.Join(c)
	c.addChannel(chid)
}

// LeaveChannel user leave a channel
func (s *Server) LeaveChannel(c *Client, chid string) {
	if chid == "*" {
		c.Ok("leave")
		for id := range c.channels {
			if ch, ok := s.Channels[id]; ok {
				ch.Leave(c)
			}
		}
	} else {
		ch, ok := s.Channels[chid]
		if !ok {
			c.Error("leave", "Channel not found", ErrChannelNotFound, false)
			return
		}
		c.Ok("leave")
		ch.Leave(c)
	}
	c.removeChannel(chid)
}

// SendMessage user sends message to channel
func (s *Server) SendMessage(c *Client, chid string, body string) {
	if body == "" {
		c.Error("message", "Body is empty", ErrEmptyField, false)
		return
	}
	if _, ok := c.channels[chid]; !ok {
		c.Error("message", "Invalid channel", ErrInvalidChannel, false)
		return
	}
	ch, ok := s.Channels[chid]
	if !ok {
		c.Error("message", "Channel not found", ErrChannelNotFound, false)
		return
	}
	c.Ok("message")

	m, err := json.Marshal(struct {
		Action string             `json:"action"`
		Data   utils.EvSrvMessage `json:"data"`
	}{
		Action: "ev_message",
		Data: utils.EvSrvMessage{
			Chid: chid,
			From: c.cid,
			Nick: c.nick,
			Body: body,
		},
	})
	if !c.CheckError(err, "Can't marhsal answer") {
		return
	}
	ch.outgoing <- m

}

///////////////// Channel Class ///////////////////////////////////////////////

// Channel is class for chat channel, keeps user informaion
type Channel struct {
	clients  map[string]*Client
	joins    chan net.Conn
	outgoing chan []byte
	name     string
	descr    string
	id       string
}

// NewChannel is construction for Channel
func NewChannel(name string, desc string) *Channel {
	c := &Channel{
		clients:  make(map[string]*Client),
		joins:    make(chan net.Conn),
		outgoing: make(chan []byte),
		name:     name,
		descr:    desc,
		id:       utils.GetMD5Hash(name),
	}
	c.Listen()
	return c
}

// Broadcast sends message to all users in channel
func (c *Channel) Broadcast(data []byte) {
	for _, client := range c.clients {
		client.outgoing <- data
	}
}

// Join adds the user to channel and send informaion about this to all users in
// channel
func (c *Channel) Join(client *Client) {
	if _, ok := c.clients[client.cid]; !ok {
		c.clients[client.cid] = client
		m, err := json.Marshal(struct {
			Action string                `json:"action"`
			Data   utils.EvSrvEnterLeave `json:"data"`
		}{
			Action: "ev_enter",
			Data: utils.EvSrvEnterLeave{
				Chid: c.id,
				Uid:  client.cid,
				Nick: client.nick,
			},
		})
		if !client.CheckError(err, "Can't marhsal answer") {
			return
		}
		c.outgoing <- m
	}
}

// Leave removes the user to channel and send informaion about this to all
// users in channel
func (c *Channel) Leave(client *Client) {
	delete(c.clients, client.cid)

	m, err := json.Marshal(struct {
		Action string                `json:"action"`
		Data   utils.EvSrvEnterLeave `json:"data"`
	}{
		Action: "ev_leave",
		Data: utils.EvSrvEnterLeave{
			Chid: c.id,
			Uid:  client.cid,
			Nick: client.nick,
		},
	})
	if !client.CheckError(err, "Can't marhsal answer") {
		return
	}
	c.outgoing <- m
}

// Listen throws data from channel outgoing to Broadcast
func (c *Channel) Listen() {
	go func() {
		for data := range c.outgoing {
			c.Broadcast(data)
		}
	}()
}
