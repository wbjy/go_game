package PoolAndAgent

import (
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/proto"
	"go_game/DialManager"
	ListenManager "go_game/TcpManager"
	bs_proto "go_game/protodefine"
	"go_game/protodefine/mytype"
	"go_game/protodefine/router"
	"go_game/protodefine/tcpnet"
	"time"
)

type NetAgent struct {
	IpAddress string
}

var netAgentList []NetAgent

func (*NetAgent) SendMsg(req *tcpnet.TCPTransferMsg) {
	var connId uint64 = req.Base.ConnId
	sess := ListenManager.GetSessionByConnId(connId)
	if sess ==nil{
		return
	}
	select {
	case sess.MsgWriteCh <- req:
		case <-time.After(5*time.Second):
			fmt.Println("严重事件，sess.MsgWriteCh已经阻塞超时5秒, connId =", connId)
	}
}

func CreateNetAgent(ipAdd string) *NetAgent {
	for _, v := range netAgentList {
		if v.IpAddress == ipAdd {
			fmt.Println("已存在相同IP地址和端口,返回空指针")
			return nil
		}
	}

	agent := new(NetAgent)
	agent.IpAddress = ipAdd
	netAgentList = append(netAgentList, *agent)
	return agent
}

type RouterAgent struct {
	IpAddress string
	DialSeesion *DialManager.ConnectSession
}

func CreateRouterAgent(ipAdd string) *RouterAgent {
	agent := new(RouterAgent)
	agent.IpAddress = ipAdd
	return agent
}

func (this *RouterAgent) RunRouterAgent(RouterToLogicChannel chan proto.Message, myAppId uint32, myAppType uint32) {
	fmt.Println("RunRouterAgent() myAppId=", myAppId, "myAppType=", myAppType)
	isClosed := false
	this.DialSeesion = DialManager.CreateClien(this.IpAddress, RouterToLogicChannel)
	for {
		//在这个router session关闭后，需要不停的尝试重新连接
		for isClosed || this.DialSeesion == nil {
			select {
			case <-time.After(700 * time.Millisecond): //每700毫秒尝试重新连接一次
			}
			this.DialSeesion = DialManager.CreateClien(this.IpAddress, RouterToLogicChannel)
			if this.DialSeesion != nil {
				fmt.Println("router Agent重连成功了")
				isClosed = false
			}
		}
		req := new(router.RegisterAppReq)
		bs_proto.SetBaseKindAndSubId(req)
		req.AppType = myAppType
		req.AppId = myAppId
		buff, err := proto.Marshal(req)
		if err != nil {
			fmt.Println("序列化RegisterAppReq出错")
		}
		tcpMsg := new(tcpnet.TCPTransferMsg)
		bs_proto.SetBaseKindAndSubId(tcpMsg)
		tcpMsg.Data = buff
		tcpMsg.DataKindId = uint32(mytype.CMDKindId_IDKindRouter)
		tcpMsg.DataSubId = uint32(router.CMDID_Router_IDRegisterAppReq)
		this.SendMsg(tcpMsg)
		fmt.Println("向router app发送了注册请求")
		//阻塞在这里直到session关闭
		select {
		case v := <- this.DialSeesion.Quit:
			if v {
				fmt.Println("router agent的连接已关闭")
				isClosed = true
				this.DialSeesion = nil
			}
		}
	}
}

func (this *RouterAgent) SendMsg(req *tcpnet.TCPTransferMsg){
	var connId uint64 = req.Base.ConnId
	if this.DialSeesion == nil { //判断是否为空
		return
	}
	//往MsgWriteCh里写req
	select{
	case this.DialSeesion.MsgWriteCh <- req:
	case <-time.After(5 * time.Second):
		fmt.Println("严重事件，DialSession.MsgWriteCh已经阻塞超时5秒, connId =", connId)
	}
}

/**
数据库相关
*/
type DBReadInfo struct {
	TheRows		*sql.Rows		//如果是读操作，这里是返回的数据集
	TheColumns	map[string]int  //返回的列名,key是列名，value是表示这个是第几列，在返回的时候好查找
	ArrValues 	[][]string		//返回的结果值
	RowNum 		int				//行数，其实就是len
}

func (this *DBReadInfo) Clear() {
	this.TheRows = nil
	this.TheColumns = nil
	this.ArrValues = nil
	this.RowNum = 0
}

type DBWriteInfo struct {
	LastId			int64	//只增列的当前值
	AffectedRows	int64	//写操作影响的行数
}

type CADODatabase struct {
	DBSourceString string		// mysql的数据库连接串 格式"root:123456@tcp(localhost:3306)/sns?charset=utf8"
	Err 			error 		//错误
	TheDB 			*sql.DB		//数据库对象
	ReadInfo		DBReadInfo	//如果是select的读操作，相关返回结果存这里
	WriteInfo		DBWriteInfo	//如果是update或insert的写操作，相关返回结果存这里
}

func CreateADODatabase(DBSourceString string) *CADODatabase {
	DBAgent := new(CADODatabase)
	DBAgent.DBSourceString = DBSourceString
	return DBAgent
}

func (this *CADODatabase) InitDB(){
	db, err := sql.Open("mysql", this.DBSourceString)
	if err != nil {
		fmt.Println("初始化失败，error=", err.Error())
		panic(err.Error())
		this.Err = err
	}
	fmt.Println("数据库连接初始化成功")
	this.TheDB = db
}

type ILogicProcess interface {
	Init(myPool *SingleMsgPool) bool							//初始化
	ProcessReq(req proto.Message, pDataBase *CADODatabase)		//处理来自所绑定的SingleMsgPool发送的报文
	OnPulse(ms uint64)											//定时函数，每隔200ms调用一次
}