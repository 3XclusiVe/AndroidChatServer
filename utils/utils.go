package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func CheckError(err error, message string, fatal bool) bool {
	if err != nil {
		fmt.Fprintln(os.Stderr, message+": "+err.Error())
		if fatal {
			os.Exit(1)
		}
	}
	return err == nil
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

//// Client messages

type CltRequest struct {
	Action  string          `json:"action"`
	RawData json.RawMessage `json:"data,omitempty"`
}

type CltAuth struct {
	Login string `json:"login"`
	Pass  string `json:"pass"`
}

type CltRegister struct {
	Nick string `json:"nick"`
	CltAuth
}

type CltBaseReq struct {
	Cid string `json:"cid"`
	Sid string `json:"sid"`
}

type CltUserInfo struct {
	User string `json:"user"`
	CltBaseReq
}

type CltSetUserInfo struct {
	UserStatus string `json:"user_status"`
	CltBaseReq
}

type CltChannel struct {
	Channel string `json:"channel"`
	CltBaseReq
}

type CltCreateChannel struct {
	Name  string `json:"name"`
	Descr string `json:"descr"`
}

type CltMessage struct {
	Body string `json:"body"`
	CltChannel
}

//// Server messages

type SrvMessage struct {
	Action  string          `json:"action"`
	Time    int             `json:"time"`
	RawData json.RawMessage `json:"data,omitempty"`
}

func Decode(r io.Reader) (x *SrvMessage, err error) {
	x = new(SrvMessage)
	if err = json.NewDecoder(r).Decode(x); err != nil {
		return
	}
	fmt.Printf("Decode\n")
	return
}

type SrvWelcomeMessage struct {
	Message string `json:"message"`
	Action  string `json:"action"`
	Time    int    `json:"time"`
}

type SrvStatusMessage struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

type SrvAddChannelMessage struct {
	ChannelID string `json:"chid"`
	SrvStatusMessage
}

type SrvStatusAuthMessage struct {
	Sid  string `json:"sid"`
	Cid  string `json:"cid"`
	Nick string `json:"nick"`
	SrvStatusMessage
}

type ChannelData struct {
	Id     string `json:"chid"`
	Name   string `json:"name"`
	Descr  string `json:"descr"`
	Online int    `json:"online"`
}

type SrvListOfChannel struct {
	Channels []ChannelData `json:"channels"`
	SrvStatusMessage
}

type SrvUserInfo struct {
	Nick       string `json:"nick"`
	UserStatus string `json:"user_status"`
	SrvStatusMessage
}

/*
{
	"action":"enter",
	"data":{
		"status":[0-9]+,
		"error":"TEXT_OF_ERROR",
		"users":[
		{
			"uid":"USER_ID",
			"nick":"NICKNAME",
		},...
		],
		"last_msg": [
		{
			"mid":"MESSAGE_ID",
			"from":"USER_ID",
			"nick":"USERS_NICKNAME",
			"body":"TEXT_OF_MESSAGE",
			"time":UNIXTIMESTAMP_OF_MESSAGE,
		}, ...
		]
	}
}
*/

type UserData struct {
	Uid  string `json:"uid"`
	Nick string `json:"nick"`
}

type MessageData struct {
	Mid  string `json:"mid"`
	From string `json:"from"`
	Nick string `json:"nick"`
	Body string `json:"body"`
	Time int    `json:"time"`
}

type SrvEnterToChannel struct {
	Users    []UserData    `json:"users"`
	LastMsgs []MessageData `json:"last_msg"`
	SrvStatusMessage
}

/*
{
	"action":"ev_enter",
	"data":{
		"chid":"CHANNEL_ID",
		"uid":"USER_ID",
		"nick":"NICKNAME"
	}
}
*/
type EvSrvEnterLeave struct {
	Chid string `json:"chid"`
	Uid  string `json:"uid"`
	Nick string `json:"nick"`
}

/*
{
	"action":"ev_message",
	"data":{
		"chid":"CHANNEL_ID",
		"from":"USER_ID",
		"nick":"NICKNAME",
		"body":"TEXT_OF_MESSAGE"
	}
}
*/
type EvSrvMessage struct {
	Chid string `json:"chid"`
	From string `json:"from"`
	Nick string `json:"nick"`
	Body string `json:"body"`
}
