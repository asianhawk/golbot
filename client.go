package main

import (
	"log"

	"github.com/yuin/gopher-lua"
)

type ChatClient interface {
	Logger() *log.Logger
	Say(target, message string)
	Connect(connspec string) error
	On(L *lua.LState, action string, fn *lua.LFunction)
	Respond(L *lua.LState, pattern string, fn *lua.LFunction)
	Serve(L *lua.LState, fn *lua.LFunction)
}

type luaChatClient struct {
	underlying *lua.LUserData
	chatClient ChatClient
}

type MessageEvent struct {
	From    string
	Target  string
	Message string
	Raw     interface{}
}

func NewMessageEvent(from, target, message string, raw interface{}) *MessageEvent {
	return &MessageEvent{from, target, message, raw}
}

func registerChatClientType(L *lua.LState, typeName string) {
	mt := L.NewTypeMetatable(typeName)
	funcs := L.SetFuncs(L.NewTable(), chatClientMethods)
	L.SetField(mt, "__index", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(2)
		if key == "raw" {
			L.Push(checkChatClientU(L))
		} else {
			L.Push(funcs.RawGetString(key))
		}
		return 1
	}))
}

func newChatClient(L *lua.LState, typeName string, chatClient ChatClient, underlyingObject *lua.LUserData) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = &luaChatClient{underlyingObject, chatClient}
	L.SetMetatable(ud, L.GetTypeMetatable(typeName))
	return ud
}

func checkChatClient(L *lua.LState) *luaChatClient {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaChatClient); ok {
		return v
	}
	L.ArgError(1, "ChatClient expected")
	return nil
}

func checkChatClientG(L *lua.LState) ChatClient {
	return checkChatClient(L).chatClient
}

func checkChatClientU(L *lua.LState) *lua.LUserData {
	return checkChatClient(L).underlying
}

var chatClientMethods = map[string]lua.LGFunction{
	"say":     chatClientSay,
	"on":      chatClientOn,
	"respond": chatClientRespond,
	"serve":   chatClientServe,
	"connect": chatClientConnect,
}

func chatClientSay(L *lua.LState) int {
	checkChatClientG(L).Say(L.CheckString(2), L.CheckString(3))
	return 0
}

func chatClientOn(L *lua.LState) int {
	checkChatClientG(L).On(L, L.CheckString(2), L.CheckFunction(3))
	return 0
}

func chatClientRespond(L *lua.LState) int {
	checkChatClientG(L).Respond(L, L.CheckString(2), L.CheckFunction(3))
	return 0
}

func chatClientServe(L *lua.LState) int {
	checkChatClientG(L).Serve(L, L.CheckFunction(2))
	return 0
}

func chatClientConnect(L *lua.LState) int {
	if err := checkChatClientG(L).Connect(L.CheckString(2)); err != nil {
		pushN(L, lua.LNil, lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}