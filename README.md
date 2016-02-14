
#Протокол работы с сервером
##Запросы от клиента на сервер
Регистрация
```
{
	"action":"register",
	"data":{
		"login":"MY_LOGIN",
		"pass":"MD5_FROM_PASS",
		"nick":"NICKNAME"
	}
}
```
Авторизация
```
{
	"action":"auth",
	"data":{
		"login":"MY_LOGIN",
		"pass":"MD5_FROM_PASS"
	}
}
```
Запрос на список каналов
```
{
	"action":"channellist",
	"data":{
		"cid":"MY_USER_ID",
		"sid":"MY_SESSION_ID"
	}
}
```
Запросить информацию о пользователе
```
{
	"action":"userinfo",
	"data": {
		"user":"USER_ID",
		"cid":"MY_USER_ID",
		"sid":"MY_SESSION_ID"
	}
}
```
Войти в канал чата
```
{
	"action":"enter",
	"data": {
		"cid":"MY_USER_ID",
		"sid":"MY_SESSION_ID",
		"channel":"NEED_CHANNEL_ID"
	}
}
```
Выйти из канала
```
{
	"action":"leave",
	"data": {
		"cid":"MY_USER_ID",
		"sid":"MY_SESSION_ID",
		"channel":"NEED_CHANNEL_ID"|"*"
	}
}
```
Отправить сообщение в канал
```
{
	"action":"message",
	"data": {
		"cid":"MY_USER_ID",
		"sid":"MY_SESSION_ID",
		"channel":"NEED_CHANNEL_ID",
		"body":"MESSAGE"
	}
}
```
Создать канал 
```
{
	"action":"createchannel",
	"data": {
		"cid":"MY_USER_ID",
		"sid":"MY_SESSION_ID",
		"name":"NAME_OF_CHANNEL",
		"descr":"DESCRIPTION_OF_CHANNEL",
	}
}
```

##Ответы сервера на клиент
Welcome сообщение приходит при конекте к серверу
```
{
	"action":"welcome",
	"message": "WELCOME_TEXT",
	"time":UNIXTIMESTAMP
}
```
Ответ на авторизацию
```
{
	"action":"auth",
	"data":{
		"status":[0-9]+,
		"error":"TEXT_OF_ERROR",
		"sid":"SESSION_ID",
		"uid":"USER_ID"
	}
}
```
Ответ на регистрацию
```
{
	"action":"register",
	"data":{
		"status":[0-9]+,
		"error":"TEXT_OF_ERROR"
	}
}
```
Ответ на запрос списка каналов чата
```
{
	"action":"channellist",
	"data":{
		"status":[0-9]+,
		"error":"TEXT_OF_ERROR",
		"channels":[
		{
			"chid":"NEED_CHANNEL_ID",
			"name":"NAME_OF_CHANNEL",
			"descr":"DESCRIPTION_OF_CHANNEL",
			"online":ONLINE_NUM,
		}, ....
		]
	}
}
```
Ответ на вход в канал (получаем пользователей и возможно несколько предыдущих сообщений)
```
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
			"time":UNIXTIMESTAMP_OF_MESSAGE
		}, ...
		]
	}
}
```
Ответ на запрос информации о пользователе
```
{
	"action":"userinfo",
	"data":{
		"status":[0-9]+,
		"error":"TEXT_OF_ERROR",
		"nick":"NICKNAME",
		"user_status":"STATUS_STRING"
	}
}
```
Ответ на выход из канала
```
{
	"action":"leave",
	"data":{
		"status":[0-9]+,
		"error":"TEXT_OF_ERROR"
	}
}
```
##События присылаемые с сервера на клиент
Кто-то вошел в канал (в том числе и вы)
```
{
	"action":"ev_enter",
	"data":{
		"chid":"CHANNEL_ID",
		"uid":"USER_ID",
		"nick":"NICKNAME"
	}
}
```
Кто-то покинул канал
```
{
	"action":"ev_leave",
	"data":{
		"chid":"CHANNEL_ID",
		"uid":"USER_ID",
		"nick":"NICKNAME"
	}
}
```
Кто-то написал сообщение в канал (в том числе и вы)
```
{
	"action":"ev_message",
	"data":{
		"chid":"CHANNEL_ID",
		"from":"USER_ID",
		"nick":"NICKNAME",
		"body":"TEXT_OF_MESSAGE"
	}
}
```

