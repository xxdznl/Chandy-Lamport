package chandy_lamport

// test_common.go 对.top .event .snap文件进行处理
import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ==================================
//  Helper methods used in test code
// ==================================

// Directory containing all the test files
const testDir = "test_data"

// Read the topology from a ".top" file.
// The expected format of the file is as follows:
//   - The first line contains number of servers N (e.g. "2")
//   - The next N lines each contains the server ID and the number of tokens on
//     that server, in the form "[serverId] [numTokens]" (e.g. "N1 1")
//   - The rest of the lines represent unidirectional links in the form "[src dst]"
//     (e.g. "N1 N2")
func readTopology(fileName string, sim *Simulator) {
	b, err := ioutil.ReadFile(path.Join(testDir, fileName))
	checkError(err)
	lines := strings.FieldsFunc(string(b), func(r rune) bool { return r == '\n' })

	// Must call this before we start logging
	//最开始的初始化logger.events
	sim.logger.NewEpoch()

	// Parse topology from lines
	numServersLeft := -1
	for _, line := range lines {
		// Ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}
		if numServersLeft < 0 {
			numServersLeft, err = strconv.Atoi(line)
			checkError(err)
			continue
		}
		// Otherwise, always expect 2 tokens
		parts := strings.Fields(line)
		if len(parts) != 2 {
			log.Fatal("Expected 2 tokens in line: ", line)
		}
		if numServersLeft > 0 {
			// This is a server
			serverId := parts[0]
			numTokens, err := strconv.Atoi(parts[1])
			checkError(err)
			sim.AddServer(serverId, numTokens)
			numServersLeft--
		} else {
			// This is a link
			src := parts[0]
			dest := parts[1]
			sim.AddForwardLink(src, dest)
		}
	}
}

// Read the events from a ".events" file and inject the events into the simulator.
// The expected format of the file is as follows:
//   - "tick N" indicates N time steps has elapsed (default N = 1)
//   - "send N1 N2 1" indicates that N1 sends 1 token to N2
//   - "snapshot N2" indicates the beginning of the snapshot process, starting on N2
//
// Note that concurrent events are indicated by the lack of ticks between the events.
// This function waits until all the snapshot processes have terminated before returning
// the snapshots collected.
// 读取.events文件内容，调用simulator逐个执行event，返回SnapshotState数组
func injectEvents(fileName string, sim *Simulator) []*SnapshotState {
	b, err := ioutil.ReadFile(path.Join(testDir, fileName))
	checkError(err)

	snapshots := make([]*SnapshotState, 0)
	//信道在使用前必须创建：仅当信道的缓冲区填满后，向其发送数据时才会阻塞。当缓冲区为空时，接受方会阻塞。
	getSnapshots := make(chan *SnapshotState, 100)
	//开启的快照命令数目
	numSnapshots := 0
	//将文件内容按行拆分成不同字符串
	lines := strings.FieldsFunc(string(b), func(r rune) bool { return r == '\n' })
	//读取每行的字符串
	for _, line := range lines {
		// Ignore comments忽略注释
		if strings.HasPrefix("#", line) {
			continue
		}
		parts := strings.Fields(line)
		//第一部分是send 还是snapshot 还是tick
		switch parts[0] {
		case "send":
			//send就读取源server号，目的server号
			src := parts[1]
			dest := parts[2]
			//token号
			tokens, err := strconv.Atoi(parts[3])
			checkError(err)
			sim.InjectEvent(PassTokenEvent{src, dest, tokens})
		case "snapshot":
			//snapshot命令数量++
			numSnapshots++
			//哪个server发起的snapshot
			serverId := parts[1]
			//此次snapshot的id
			snapshotId := sim.nextSnapshotId
			//sim执行snapshotevent事件
			sim.InjectEvent(SnapshotEvent{serverId})
			//由 Go 运行时管理的轻量级线程检测算法是否结束，
			//不断重复运行，直到所有server都结束后getSnapshots通道才有数据
			//注意，快照开始后只要events没执行完不影响程序的进行，因此有几次snapshot命令就会产生几个.snap文件
			go func(id int) {
				//将所有server结束时的状态写入getSnapshots通道
				getSnapshots <- sim.CollectSnapshot(id)
			}(snapshotId)
		//每tick一次，处理每个server的发送通道上的一条信息（如果有的话）
		case "tick":
			numTicks := 1       //默认没有第二个参数，就tick 1次
			if len(parts) > 1 { //如果有第二个参数  tick 4 10等
				//字符串转换为整数
				numTicks, err = strconv.Atoi(parts[1])
				checkError(err)
			}
			//至少tick 1次
			for i := 0; i < numTicks; i++ {
				sim.Tick()
			}
		default:
			log.Fatal("Unknown event command: ", parts[0])
		}
	}
	//所有事件执行完毕后，如果numSnapshot >0 即通道上还有快照命令没执行完
	//执行完毕后将调用sim.CollectSnapshot(id)方法获得的全局状态数据取出
	// Keep ticking until snapshots complete
	for numSnapshots > 0 {
		select {
		//一条快照指令执行完毕，通道上产生了数据
		case snap := <-getSnapshots:
			//fmt.Println("_________________111111111111__________")
			//fmt.Println("snap:", snap)
			snapshots = append(snapshots, snap)
			numSnapshots--
		//还有快照指令没执行 继续tick
		default:
			sim.Tick()
		}
	}
	//还要考虑最后一批消息的延迟接收时间
	//即便server结束，位于接收通道上的未接收的消息也要进行记录
	// Keep ticking until we're sure that the last message has been delivered
	for i := 0; i < maxDelay+1; i++ {
		sim.Tick()
	}

	return snapshots
}

// Read the state of snapshot from a ".snap" file.
// The expected format of the file is as follows:
//   - The first line contains the snapshot ID (e.g. "0")
//   - The next N lines contains the server ID and the number of tokens on that server,
//     in the form "[serverId] [numTokens]" (e.g. "N1 0"), one line per server
//   - The rest of the lines represent messages exchanged between the servers,
//     in the form "[src] [dest] [message]" (e.g. "N1 N2 token(1)")
func readSnapshot(fileName string) *SnapshotState {
	b, err := ioutil.ReadFile(path.Join(testDir, fileName))
	checkError(err)
	snapshot := SnapshotState{0, make(map[string]int), make([]*SnapshotMessage, 0)}
	lines := strings.FieldsFunc(string(b), func(r rune) bool { return r == '\n' })
	for _, line := range lines {
		// Ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 1 {
			// Snapshot ID
			snapshot.id, err = strconv.Atoi(line)
			checkError(err)
		} else if len(parts) == 2 {
			// Server and its tokens
			serverId := parts[0]
			numTokens, err := strconv.Atoi(parts[1])
			checkError(err)
			snapshot.tokens[serverId] = numTokens
		} else if len(parts) == 3 {
			// Src, dest and message
			src := parts[0]
			dest := parts[1]
			messageString := parts[2]
			var message interface{}
			if strings.Contains(messageString, "token") {
				pattern := regexp.MustCompile(`[0-9]+`)
				matches := pattern.FindStringSubmatch(messageString)
				if len(matches) != 1 {
					log.Fatal("Unable to parse token message: ", messageString)
				}
				numTokens, err := strconv.Atoi(matches[0])
				checkError(err)
				message = TokenMessage{numTokens}
			} else {
				log.Fatal("Unknown message: ", messageString)
			}
			snapshot.messages =
				append(snapshot.messages, &SnapshotMessage{src, dest, message})
		}
	}
	return &snapshot
}

// Helper function to pretty print the tokens in the given snapshot state
func tokensString(tokens map[string]int, prefix string) string {
	str := make([]string, 0)
	for _, serverId := range GetSortedKeys(tokens) {
		numTokens := tokens[serverId]
		maybeS := "s"
		if numTokens == 1 {
			maybeS = ""
		}
		str = append(str, fmt.Sprintf(
			"%v%v: %v token%v", prefix, serverId, numTokens, maybeS))
	}
	return strings.Join(str, "\n")
}

// Helper function to pretty print the messages in the given snapshot state
func messagesString(messages []*SnapshotMessage, prefix string) string {
	str := make([]string, 0)
	for _, msg := range messages {
		str = append(str, fmt.Sprintf(
			"%v%v -> %v: %v", prefix, msg.src, msg.dest, msg.message))
	}
	return strings.Join(str, "\n")
}

// Assert that the two snapshot states are equal.
// If they are not equal, throw an error with a helpful message.
func assertEqual(expected, actual *SnapshotState) {
	if expected.id != actual.id {
		log.Fatalf("Snapshot IDs do not match: %v != %v\n", expected.id, actual.id)
	}
	if len(expected.tokens) != len(actual.tokens) {
		log.Fatalf(
			"Snapshot %v: Number of tokens do not match."+
				"\nExpected:\n%v\nActual:\n%v\n",
			expected.id,
			tokensString(expected.tokens, "\t"),
			tokensString(actual.tokens, "\t"))
	}
	if len(expected.messages) != len(actual.messages) {
		log.Fatalf(
			"Snapshot %v: Number of messages do not match."+
				"\nExpected:\n%v\nActual:\n%v\n",
			expected.id,
			messagesString(expected.messages, "\t"),
			messagesString(actual.messages, "\t"))
	}
	for id, tok := range expected.tokens {
		if actual.tokens[id] != tok {
			log.Fatalf(
				"Snapshot %v: Tokens on %v do not match."+
					"\nExpected:\n%v\nActual:\n%v\n",
				expected.id,
				id,
				tokensString(expected.tokens, "\t"),
				tokensString(actual.tokens, "\t"))
		}
	}
	// Ensure message order is preserved per destination
	// Note that we don't require ordering of messages across all servers to match
	// message不用按序 map[string][] 第i个server的第j个消息，先把所有消息处理成这种格式，方便比较
	expectedMessages := make(map[string][]*SnapshotMessage)
	actualMessages := make(map[string][]*SnapshotMessage)
	for i := 0; i < len(expected.messages); i++ {
		em := expected.messages[i]
		am := actual.messages[i]
		_, ok1 := expectedMessages[em.dest]
		_, ok2 := actualMessages[am.dest]
		if !ok1 {
			expectedMessages[em.dest] = make([]*SnapshotMessage, 0)
		}
		if !ok2 {
			actualMessages[am.dest] = make([]*SnapshotMessage, 0)
		}
		expectedMessages[em.dest] = append(expectedMessages[em.dest], em)
		actualMessages[am.dest] = append(actualMessages[am.dest], am)
	}
	// Test message order per destination
	// map的值一样但顺序不一样，  m1:=map[string]int{"a":1,"b":2,"c":3};
	//							m2:=map[string]int{"a":1,"c":3,"b":2};
	//使用reflect.DeepEqual进行比较
	for dest := range expectedMessages {
		ems := expectedMessages[dest]
		ams := actualMessages[dest]
		if !reflect.DeepEqual(ems, ams) {
			log.Fatalf(
				"Snapshot %v: Messages received at %v do not match."+
					"\nExpected:\n%v\nActual:\n%v\n",
				expected.id,
				dest,
				messagesString(ems, "\t"),
				messagesString(ams, "\t"))
		}
	}
}

// Helper function to sort the snapshot states by ID.
func sortSnapshots(snaps []*SnapshotState) {
	sort.Slice(snaps, func(i, j int) bool {
		s1 := snaps[i]
		s2 := snaps[j]
		return s2.id > s1.id
	})
}

// Verify that the total number of tokens recorded in the snapshot preserves
// the number of tokens in the system
// 系统内token如何交换应该不影响总的token数目
func checkTokens(sim *Simulator, snapshots []*SnapshotState) {
	expectedTokens := 0
	//系统中所有server共有多少token
	for _, server := range sim.servers {
		expectedTokens += server.Tokens
	}
	//不论进行到第几次snapshot，每轮的snapshot都应保持总量不变
	for _, snap := range snapshots {
		snapTokens := 0
		// Add tokens recorded on servers
		for _, tok := range snap.tokens {
			snapTokens += tok
		}
		// Add tokens from messages in-flight
		// 接收通道内未取出的消息token也要加上
		for _, message := range snap.messages {
			switch msg := message.message.(type) {
			case TokenMessage:
				snapTokens += msg.numTokens
			}
		}
		if expectedTokens != snapTokens {
			log.Fatalf("Snapshot %v: simulator has %v tokens, snapshot has %v:\n%v\n%v",
				snap.id,
				expectedTokens,
				snapTokens,
				tokensString(snap.tokens, "\t"),
				messagesString(snap.messages, "\t"))
		}
	}
}
