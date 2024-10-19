package chandy_lamport

import (
	"fmt"
	"log"
	"reflect"
	"sort"
)

const debug = true
// ====================================
//  Messages exchanged between servers
// ====================================

// An event that represents the sending of a message.
// This is expected to be queued in `link.events`.
type SendMessageEvent struct {
	src     string
	dest    string
	message interface{} //发送的事件可能包含 token信息，或者marker信息
	// The message will be received by the server at or after this time step
	receiveTime int //接收时刻，只会更晚
}

//所有信息应该被封装在sendMessageEvent事件中

// A message sent from one server to another for token passing.
// This is expected to be encapsulated within a `sendMessageEvent`.
// token信息，成员变量为token数目
type TokenMessage struct {
	numTokens int
}

// 输出token信息内容
func (m TokenMessage) String() string {
	return fmt.Sprintf("token(%v)", m.numTokens)
}

// A message sent from one server to another during the chandy-lamport algorithm.
// This is expected to be encapsulated within a `sendMessageEvent`.
// marker信息 ，成员变量为snapshotId ？？？？
type MarkerMessage struct {
	snapshotId int
}

// 打印marker信息内容
func (m MarkerMessage) String() string {
	return fmt.Sprintf("marker(%v)", m.snapshotId)
}

// =======================
//  Events used by logger
// =======================

// A message that signifies receiving of a message on a particular server
// This is used only for debugging that is not sent between servers
type ReceivedMessageEvent struct {
	src     string
	dest    string
	message interface{}
}

func (m ReceivedMessageEvent) String() string {
	switch msg := m.message.(type) {
	case TokenMessage:
		return fmt.Sprintf("%v received %v tokens from %v", m.dest, msg.numTokens, m.src)
	case MarkerMessage:
		return fmt.Sprintf("%v received marker(%v) from %v", m.dest, msg.snapshotId, m.src)
	}
	return fmt.Sprintf("Unrecognized message: %v", m.message)
}

// A message that signifies sending of a message on a particular server
// This is used only for debugging that is not sent between servers
// 发送信息事件
type SentMessageEvent struct {
	src     string
	dest    string
	message interface{} //可能为TokenMessage 可能为MarkerMessage
}

func (m SentMessageEvent) String() string {
	switch msg := m.message.(type) {
	case TokenMessage:
		return fmt.Sprintf("%v sent %v tokens to %v", m.src, msg.numTokens, m.dest)
	case MarkerMessage:
		return fmt.Sprintf("%v sent marker(%v) to %v", m.src, msg.snapshotId, m.dest)
	}
	return fmt.Sprintf("Unrecognized message: %v", m.message)
}

// A message that signifies the beginning of the snapshot process on a particular server.
// This is used only for debugging that is not sent between servers.
type StartSnapshot struct {
	serverId   string
	snapshotId int
}

func (m StartSnapshot) String() string {
	return fmt.Sprintf("%v startSnapshot(%v)", m.serverId, m.snapshotId)
}

// A message that signifies the end of the snapshot process on a particular server.
// This is used only for debugging that is not sent between servers.
type EndSnapshot struct {
	serverId   string
	snapshotId int
}

func (m EndSnapshot) String() string {
	return fmt.Sprintf("%v endSnapshot(%v)", m.serverId, m.snapshotId)
}

// ================================================
//  Events injected to the system by the simulator
// ================================================

// An event parsed from the .event files that represent the passing of tokens
// from one server to another
// token传递事件
type PassTokenEvent struct {
	src    string
	dest   string
	tokens int
}

// An event parsed from the .event files that represent the initiation of the
// chandy-lamport snapshot algorithm
// snapshot开始事件
type SnapshotEvent struct {
	serverId string
}

// A message recorded during the snapshot process
// snapshot进程开始后记录的消息
type SnapshotMessage struct {
	src     string
	dest    string
	message interface{}
}

// State recorded during the snapshot process
// snapshot进程开始后记录的状态
type SnapshotState struct {
	id int
	//各个server的最终token数目
	tokens map[string]int // key = server ID, value = num tokens
	//快照开始后，又记录的消息
	messages []*SnapshotMessage
}

// =====================
//  Misc helper methods
// =====================

// If the error is not nil, terminate
// 是否为nil
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Return the keys of the given map in sorted order.
// Note: The argument passed in MUST be a map, otherwise an error will be thrown.
// 获取map的keys
func GetSortedKeys(m interface{}) []string {
	//通过反射获取map
	v := reflect.ValueOf(m)
	//kind()获取是基础类型  type MyInt int中 Name()获取MyInt kind()获取int
	if v.Kind() != reflect.Map {
		log.Fatal("Attempted to access sorted keys of a non-map: ", m)
	}
	//初始化string[]数组
	keys := make([]string, 0)
	//reflect.MapKeys()函数用于获取一个切片，该切片包含未指定顺序的Map中存在的所有键
	for _, k := range v.MapKeys() {
		keys = append(keys, k.String())
	}
	//返回所有键
	sort.Strings(keys)
	return keys
}
