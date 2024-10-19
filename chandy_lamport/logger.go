package chandy_lamport

import (
	"fmt"
	"log"
)

// =================================
//  Event logger, internal use only
// =================================

type Logger struct {
	// index = time step
	// value = events that occurred at that time step
	events [][]LogEvent
}

type LogEvent struct {
	//server号
	serverId string
	// Number of tokens before execution of event
	//在事件发生前server的token数
	serverTokens int
	event        interface{}
}

// 重写String方法根据事件类型，打印事件
func (event LogEvent) String() string {
	//预置WithTokens
	prependWithTokens := false
	switch evt := event.event.(type) {
	//当前事件为发送事件
	case SentMessageEvent:
		switch evt.message.(type) {
		case TokenMessage:
			prependWithTokens = true
		}
		//当前事件为接收事件
	case ReceivedMessageEvent:
		switch evt.message.(type) {
		case TokenMessage:
			prependWithTokens = true
		}
		//当前事件为开始算法流程事件
	case StartSnapshot:
		prependWithTokens = true
		//结束算法事件才为false
	case EndSnapshot:
	default:
		log.Fatal("Attempted to log unrecognized event: ", event.event)
	}
	if prependWithTokens {
		return fmt.Sprintf("%v has %v token(s)\n\t%v",
			event.serverId,
			event.serverTokens,
			event.event)
	} else {
		return fmt.Sprintf("%v", event.event)
	}
}

// 新建logger
func NewLogger() *Logger {
	return &Logger{make([][]LogEvent, 0)}
}

// 格式化输出logger的事件
func (log *Logger) PrettyPrint() {
	for epoch, events := range log.events {
		//如果当前轮次事件不为空，输出当前时刻
		if len(events) != 0 {
			fmt.Printf("Time %v:\n", epoch)
		}
		//当前时刻发生的事件，逐个输出
		for _, event := range events {
			fmt.Printf("\t%v\n", event)
		}
	}
}

// 新一时刻，向 log.events二维数组上，新增一层[]logEvent
func (log *Logger) NewEpoch() {
	log.events = append(log.events, make([]LogEvent, 0))
}

// 记录事件
func (logger *Logger) RecordEvent(server *Server, event interface{}) {
	//上一次记录的事件的层数 比如 a[2][2] len(a) = 2 要访问a[1]需要再减一
	mostRecent := len(logger.events) - 1
	//获得该层 注意是临时拷贝出的，在events上改不影响 logger.events
	events := logger.events[mostRecent]
	//向该层添加新事件
	events = append(events, LogEvent{server.Id, server.Tokens, event})
	//写回logger.events
	logger.events[mostRecent] = events
}
