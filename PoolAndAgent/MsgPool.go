package PoolAndAgent

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	ListenManager "go_game/TcpManager"
	bs_proto "go_game/protodefine"
	"go_game/protodefine/mytype"
	"go_game/protodefine/router"
	"go_game/protodefine/tcpnet"
	"sync/atomic"
	"time"
)

const CHANNEL_LENGTH = 100000 //消息管道的容量
const ONPULSE_INTERVAL = 100  //LogicProcess的OnPulse函数定时调用的时间间隔，单位毫秒

type SingleMsgPool struct {
	quit chan int // 结束程序运行报文，传输的是appid
	IsInit bool
	NetToLogicChannel chan proto.Message //网络层写Pool层读后往楼基层发送的消息管道
	bindingNetAgent *NetAgent    	//与pool绑定的NetAgent网络层
	//与router相连的网络层，bindingRouterAgent除了router app本身没有，
	//其他app都有，因为其他app的消息都需要router去转发
	RouterToLogicChannel chan proto.Message
	bindingRouterAgent	 *RouterAgent
	//数据库层
	bindingDBProcess [](*CADODatabase)
	PoolToLogicChannel chan proto.Message		//Pool层写逻辑层度的消息管道
	bindingLogicProcesses	[]ILogicProcess		//与pool绑定的逻辑处理层
	myAPPType			uint32
	myAppId				uint32
}

func CreateMsgPool(quit chan int, myAppType uint32, myAppId uint32) *SingleMsgPool {
	pool := new(SingleMsgPool)
	pool.IsInit = false
	pool.bindingNetAgent = nil
	pool.quit = quit
	pool.myAPPType = myAppType
	pool.myAppId = myAppId
	bs_proto.OutputMyLog("CreateMsgPool() myAppType=",myAppType)
	return pool
}

func (this *SingleMsgPool) StopRun() {
	if this.quit != nil {
		this.quit <- 1
	}
}


//this并不是golang关键字
/*初始化并运行
//初始化并运行，如果bindingLogicProcesses的len大于1的话，而且需要对每个ILogicProcess都进行额外的初始化，那么InitMsg这个就是初始化报文
//比如数据库处理的pool就需要这么初始化，因为数据库的IO比较慢，所以需要多协程。
//假设我开了10个数据库协程，如果我不一开始把InitMsg推给每个ILogicProcess都处理的话，后面ILogicProcess就无法全部都初始化了
//因为10个ILogicProcess协程是共同读取同一个PoolToLogicChannel管道，所以如果像主逻辑协程那样在初始化后再发送初始化报文，那么如果连续发送10个报文
//可能是有些协程重复收到了初始化报文，有些协程没有收到初始化报文，所以一定要在go RunLogicProcess把每个协程ILogicProcess都初始化
*/
func (this *SingleMsgPool) InitAndRun(InitMsg proto.Message) bool {
	if len(this.bindingLogicProcesses) == 0 {
		fmt.Println("pool必须绑定逻辑处理实例")
		return false
	}

	this.PoolToLogicChannel = make(chan proto.Message, CHANNEL_LENGTH)
	if this.bindingNetAgent != nil {
		this.NetToLogicChannel = make(chan proto.Message, CHANNEL_LENGTH)
		//创建服务端监听协程
		go ListenManager.CreateServer(this.bindingNetAgent.IpAddress, this.NetToLogicChannel)
	}
	if this.bindingRouterAgent != nil {
		this.RouterToLogicChannel = make(chan proto.Message, CHANNEL_LENGTH)
		go this.bindingRouterAgent.RunRouterAgent(this.RouterToLogicChannel, this.myAppId, this.myAPPType)
	}
	if this.NetToLogicChannel != nil {
		go func() {
			for {
				select {
				case v := <- this.NetToLogicChannel:
					switch data := v.(type) {
					case *tcpnet.TCPSessionKick:
						sess := ListenManager.GetSessionByConnId(data.Base.ConnId)
						if sess != nil && atomic.LoadInt32(&(sess.IsSendKickMsg)) == 0 {//原子操作，判断sess.IsSendKickMsg == 0
							go sess.CloseSession(this.NetToLogicChannel) //关闭session
						}
					case *tcpnet.TCPSessionCome:
						fmt.Println("收到了TCPSessionCome报文，base=", data.Base)
						fmt.Println("NetToLogicChannel转发给PoolToLogicChannel")
						select {
						case this.PoolToLogicChannel <- v:
							case <- time.After(5 * time.Second):
								fmt.Println("严重事件，PoolToLogicChannel已经阻塞超时5秒")
						}
					default:
						fmt.Println("NetToLogicChannel转发给PoolToLogicChannel")
						select {
						case this.PoolToLogicChannel <- v:
						case <- time.After(5 * time.Second):
							fmt.Println("严重事件，PoolToLogicChannel已经阻塞超时5秒")
						}
					}
				}
			}
		}()
	}
	if this.RouterToLogicChannel != nil {
		go func() {
			select {
			case v := <- this.RouterToLogicChannel:
				switch data := v.(type) {
				case *tcpnet.TCPSessionKick:
					fmt.Println("收到了无意义的TCPSessionKick报文")
				case *tcpnet.TCPSessionCome:
					fmt.Println("收到了无意义的TCPSessionCome报文")
				case *tcpnet.TCPSessionClose:
					fmt.Println("收到了无意义的TCPSessionCome报文")
				case *tcpnet.TCPTransferMsg:
					if data.DataKindId != uint32(mytype.CMDKindId_IDKindRouter){
						break
					}
					var pRouterTran *router.RouterTransferData = nil
					switch data.DataSubId {
					case uint32(router.CMDID_Router_IDRegisterAppRsp):
						msg := new(router.RegisterAppRsp)
						err := proto.Unmarshal(data.Data, msg)
						if err != nil {
							break
						}
						fmt.Println("收到注册回复，RegResult=", msg.RegResult)
					case uint32(router.CMDID_Router_IDTransferData):
						pRouterTran = new(router.RouterTransferData)
						err := proto.Unmarshal(data.Data, pRouterTran)
						if err != nil {
							break
						}
					}
					if pRouterTran == nil {
						break
					}
					//写RouterTransferData到PoolToLogicChannel中
					fmt.Println("RouterToLogicChannel转发给PoolToLogicChannel")
					select {
					case this.PoolToLogicChannel <- pRouterTran:
						case <-time.After(5 * time.Second):
							fmt.Println("严重事件，PoolToLogicChannel已经阻塞超时5秒")
					}
				default:
					fmt.Println("收到了未知报文")
				}
			}
		}()
	}

	//根据需要创建的协程数，创建不停的读取PoolToLogicChannel的协程，
	//在执行了init函数后，循环执行ProcessReq()函数和OnPluse()定时函数
	for i, logic := range this.bindingLogicProcesses {
		var pDB *CADODatabase = nil
		if len(this.bindingLogicProcesses) != 0 && len(this.bindingDBProcess) == len(this.bindingLogicProcesses){
			pDB = this.bindingDBProcess[i]
		}
		go this.RunLogicProcess(logic, pDB, InitMsg)
	}
	this.IsInit = true
	return true
}

func (this *SingleMsgPool) RunLogicProcess(pLogic ILogicProcess, pDataBase *CADODatabase, InitMsg proto.Message){
	pLogic.Init(this)
	if InitMsg != nil {
		pLogic.ProcessReq(InitMsg, nil)
	}
	t1 := time.NewTimer(ONPULSE_INTERVAL*time.Millisecond)
	var nMs uint64 = uint64(ONPULSE_INTERVAL)
	for {
		select {
		case v, ok := <- this.PoolToLogicChannel:
			if ok {
				pLogic.ProcessReq(v, pDataBase)
			}
			case <- t1.C:
				pLogic.OnPulse(nMs)
				t1.Reset(ONPULSE_INTERVAL*time.Millisecond)
				nMs += uint64(ONPULSE_INTERVAL)
		}
	}
}

func (this *SingleMsgPool) BindNetAgent(agent *NetAgent) {
	this.bindingNetAgent = agent
}

func (this *SingleMsgPool) BindRouterAgent(agent *RouterAgent) {
	this.bindingRouterAgent = agent
}

func (this *SingleMsgPool) AddDataBaseProcess(process *CADODatabase) {
	this.bindingDBProcess = append(this.bindingDBProcess, process)
	process.InitDB() //这个InitDB并没有去连接数据库
}

func (this *SingleMsgPool) AddLogicProcess(agent ILogicProcess) {
	this.bindingLogicProcesses = append(this.bindingLogicProcesses, agent)
}

func (this *SingleMsgPool) PushMsg(req proto.Message, nMs uint64) {
	//因为在实际业务中需要延时发送的报文不多所以如果有延时发送另起一个协程
	if nMs != 0 {
		go func(delayTime uint64) {
			select {
			//阻塞delayTime * time.Millisecond后向PoolToLogicChannel发送req
			case <-time.After(time.Duration(delayTime) * time.Millisecond):
				fmt.Println("业务逻辑延时发给PoolToLogicChannel")
				this.PoolToLogicChannel <- req //因为是起了一个协程来发送，所以这里就算阻塞了也没事，就不判断超时事件了
			}
		}(nMs)
	} else { //立即推送进channel
		fmt.Println("业务逻辑直接立即发给PoolToLogicChannel")
		select {
		case this.PoolToLogicChannel <- req:
		case <-time.After(5 * time.Second):
			fmt.Println("严重事件，PoolToLogicChannel已经阻塞超时5秒")
		}
	}
}

func (this *SingleMsgPool) SendMsgToClientByNetAgent(req proto.Message) {
	if this.bindingNetAgent != nil {
		switch data := req.(type) {
		case *tcpnet.TCPTransferMsg:
			this.bindingNetAgent.SendMsg(data)
		case *tcpnet.TCPSessionKick:
			sess := ListenManager.GetSessionByConnId(data.Base.ConnId)
			if sess != nil && atomic.LoadInt32(&(sess.IsSendKickMsg)) == 0 { //原子操作
				go sess.CloseSession(this.NetToLogicChannel) //关闭此session
			}
		default:
			fmt.Println("不处理TCPSessionKick和TCPTransferMsg以外的报文")
		}
	}
}

func (this *SingleMsgPool) SendMsgToServerAppByRouter(req proto.Message) {
	if this.bindingRouterAgent != nil {
		switch data := req.(type) {
		case *tcpnet.TCPTransferMsg:
			this.bindingRouterAgent.SendMsg(data)
		case *router.RouterTransferData:
			//先把RouterTransferData转成TCPTransferMsg再发送
			buff, err := proto.Marshal(data)
			if err != nil {
				break
			}
			tcpTran := new(tcpnet.TCPTransferMsg)
			bs_proto.SetBaseKindAndSubId(tcpTran)
			bs_proto.CopyBaseExceptKindAndSubId(tcpTran.Base, data.Base)
			tcpTran.Data = buff
			tcpTran.DataKindId = uint32(mytype.CMDKindId_IDKindRouter)
			tcpTran.DataSubId = uint32(router.CMDID_Router_IDTransferData)
			this.bindingRouterAgent.SendMsg(tcpTran)
		default:
			fmt.Println("不处理TCPTransferMsg和RouterTransferData以外的报文")
		}
	}
}