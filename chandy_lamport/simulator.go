package chandy_lamport

//Simulator 是整个框架的核心，其负责
//初始化 server
//初始化网络拓扑结构
//驱动事件发生
import (
	"log"
	"math/rand"
	//为随机数添加当前时刻初始化种子
)

// Max random delay added to packet delivery
// 随机时延，模拟从通道上接收一条信息所额外花费的时间
const maxDelay = 5

// Simulator is the entry point to the distributed snapshot application.
// Simulator是该快照应用的 entry point
//
// It is a discrete time simulator, i.e. events that happen at time t + 1 come
// strictly after events that happen at time t. At each time step, the simulator
// examines messages queued up across all the links in the system and decides
// which ones to deliver to the destination.
// 离散的时间模拟器，例如t+1时刻发生的事件严格再t时刻发送的事件之后，每个时间步内，模拟器检查
// 系统内的所有通道中排队的信息，之后决定哪一个信息发送到目的server
//
// The simulator is responsible for starting the snapshot process, inducing servers
// to pass tokens to each other, and collecting the snapshot state after the process
// has terminated.
// 模拟器负责开始整个snapshot进程，包括server间互相传递货物，并在进程结束后收集快照状态
type Simulator struct {
	time           int                //当前时间
	nextSnapshotId int                //当前快照的id
	servers        map[string]*Server // key = server ID //所有server结点id
	logger         *Logger            //记录器，用于记录发送信息事件的日志
	// TODO: ADD MORE FIELDS HERE
	isSnapshotIdEnd *SyncMapIntInt //第i个Snapshot是否结束
}

// var mutexForisSnapshotIdEnd sync.Mutex //go协程互斥访问isSnapshotIdEnd map[int]int
// 新建simulator
func NewSimulator() *Simulator {
	return &Simulator{
		0,
		0,
		make(map[string]*Server),
		NewLogger(),
		NewSyncMapIntInt(),
	}
}

// Return the receive time of a message after adding a random delay.
// Note: since we only deliver one message to a given server at each time step,
// the message may be received *after* the time step returned in this function.
// 返回加入随机延迟后的信息接收时刻
// 注意由于每1个时间只传递一条信息，信息可能在这个函数返回的时间值之后才被接收到，
// 例如返回接收时刻后，并没有选择接收当前信息
func (sim *Simulator) GetReceiveTime() int {
	//rand.Seed(time.Now().UnixNano())   //当前时刻作为随机种子snapshot_test已经初始化种子
	return sim.time + 1 + rand.Intn(5) //当前时间 + 1 （必需花费1时间） + 随机时延
}

// Add a server to this simulator with the specified number of starting tokens
// 为simulator 添加一个server结点
func (sim *Simulator) AddServer(id string, tokens int) {
	server := NewServer(id, tokens, sim)
	sim.servers[id] = server
}

// Add a unidirectional link between two servers
// 为server两两间建立通道
func (sim *Simulator) AddForwardLink(src string, dest string) {
	server1, ok1 := sim.servers[src]
	server2, ok2 := sim.servers[dest]
	if !ok1 {
		log.Fatalf("Server %v does not exist\n", src)
	}
	if !ok2 {
		log.Fatalf("Server %v does not exist\n", dest)
	}
	//只需为源server建立发送通道，该通道即为目的server的接收通道
	server1.AddOutboundLink(server2)
}

// Run an event in the system
// 根据event类型运行一个event （所有server发送信息，首先记录成一个event，然后根据event类型进行不同的处理）
// 读取文件中的events后开始逐个event执行
func (sim *Simulator) InjectEvent(event interface{}) {
	switch event := event.(type) {
	case PassTokenEvent: //传递token event
		src := sim.servers[event.src] //获取源server
		//源server发送token 到目的server
		src.SendTokens(event.tokens, event.dest)
	case SnapshotEvent: //从指定server开始snapshot过程
		sim.StartSnapshot(event.serverId)
	default:
		log.Fatal("Error unknown event: ", event)
	}
}

// Advance the simulator time forward by one step, handling all send message events
// that expire at the new time step, if any.
// 推进sim时钟，处理所有server发送通道上的消息（a->b的a发送通道即为 b的接收通道，只需遍历所有发送通道进行消息处理即可）
func (sim *Simulator) Tick() {
	sim.time++
	//初始化下一个时刻的logger.events
	sim.logger.NewEpoch()
	// Note: to ensure deterministic ordering of packet delivery across the servers,
	// we must also iterate through the servers and the links in a deterministic way
	//按所有server序号顺序处理消息event
	for _, serverId := range GetSortedKeys(sim.servers) {
		server := sim.servers[serverId]
		//对当前server的所有发送通道（对于目的server来说，此时的通道为接收通道）按序处理消息
		for _, dest := range GetSortedKeys(server.outboundLinks) {
			//获取当前server的发送通道，
			link := server.outboundLinks[dest]
			// Deliver at most one packet per server at each time step to
			// establish total ordering of packet delivery to each server
			//一个时间步最多只处理通道上的一个消息
			//通道上有事件未处理
			if !link.events.Empty() {
				//取出头部 发送 事件,但不从队列删除，因为当前时刻并不一定是到货时间
				//link上只有sendMessageEvent？
				e := link.events.Peek().(SendMessageEvent)
				if e.receiveTime <= sim.time {
					//如果当前时钟时刻，大于该消息的接收时刻，就取出通道上的事件
					link.events.Pop()
					//logger记录取出消息时的事件
					sim.logger.RecordEvent(
						sim.servers[e.dest],
						ReceivedMessageEvent{e.src, e.dest, e.message})
					//todo 目的server对消息进行处理
					sim.servers[e.dest].HandlePacket(e.src, e.message)
					break
				}
			}
		}
	}
}

// Start a new snapshot process at the specified server
// sim启动一次snapshot在serverId服务器上
func (sim *Simulator) StartSnapshot(serverId string) {
	snapshotId := sim.nextSnapshotId
	sim.nextSnapshotId++
	sim.logger.RecordEvent(sim.servers[serverId], StartSnapshot{serverId, snapshotId})
	// TODO: IMPLEMENT ME
	//调用编号为serverId的server的StartSnapshot方法，发起marker
	sim.servers[serverId].StartSnapshot(snapshotId)
}

// Callback for servers to notify the simulator that the snapshot process has
// completed on a particular server
// 通知simulator一次snapshot在server i上结束
// sim可以直接通过访问server.markerFlagAt_ithSnapshot.MarkerNum判断该server上的snapshot是否结束
func (sim *Simulator) NotifySnapshotComplete(serverId string, snapshotId int) {
	//记录当前server结束，事件为EndSnapshot
	sim.logger.RecordEvent(sim.servers[serverId], EndSnapshot{serverId, snapshotId})
	// TODO: IMPLEMENT ME
	// 检测其它server是否完成snapshot
	flag := 1 //默认完成
	for _, serveri := range GetSortedKeys(sim.servers) {
		value, ok := sim.servers[serveri].markerFlagAt_ithSnapshot.Load(snapshotId)
		var temp *LogStateAndTestMarker
		if value != nil {
			temp = value.(*LogStateAndTestMarker)
		}
		//marker还没到，相关数据结构还没初始化，则肯定没完成snapshot
		if !ok {
			flag = 0
			break
			//如果MarkerNum !=邻居的接收通道数目，即还没接收到其余n-1个marker
		} else if temp.MarkerNum != (len(sim.servers[serveri].inboundLinks)) {
			flag = 0
			break
		}
	}
	//都完成了置一个完成标记
	if flag == 1 {
		sim.isSnapshotIdEnd.Store(snapshotId, 1)
	}
}

// Collect and merge snapshot state from all the servers.
// This function blocks until the snapshot process has completed on all servers.
// 收集并合并所有server的 snapshot state，直到一条snapshot命令执行完毕后才会调用该方法
// 一条snapshot命令执行完毕：每个server都发起或收到一条marker消息，且从每个接收通道都收到marker
func (sim *Simulator) CollectSnapshot(snapshotId int) *SnapshotState {
	snap := SnapshotState{snapshotId, make(map[string]int), make([]*SnapshotMessage, 0)}
	// TODO: IMPLEMENT ME
	// 判断某一snapshot命令是否执行完毕
	for {
		flagSnapshotI, _ := sim.isSnapshotIdEnd.Load(snapshotId)
		if flagSnapshotI == 1 {
			//执行完毕后返回全局状态
			//一次snapshot结束后，取出这个过程中发生的event
			for _, serveri := range GetSortedKeys(sim.servers) {
				//获取当前server的token状态
				value, _ := sim.servers[serveri].markerFlagAt_ithSnapshot.Load(snapshotId)
				temp := value.(*LogStateAndTestMarker)
				snap.tokens[serveri] = temp.Tokens
				//获取当前server的message状态
				for _, srcServer := range GetSortedKeys(temp.MessageQueueBeforeMarker) {
					for !temp.MessageQueueBeforeMarker[srcServer].Empty() {
						e := temp.MessageQueueBeforeMarker[srcServer].Peek().(*SnapshotMessage)
						temp.MessageQueueBeforeMarker[srcServer].Pop()
						snap.messages = append(snap.messages, e)
					}
				}
			}
			break
		}
	}
	return &snap
}
