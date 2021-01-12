package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"go_game/PoolAndAgent"
	bs_proto "go_game/protodefine"
	"go_game/protodefine/mytype"
	"go_game/protodefine/router"
	"go_game/protodefine/tcpnet"
	"log"
	"math/rand"
	"time"
)

type RouterConnection struct {
	connId 			uint64
	isConnectted 	bool
	clientAdress	string
	appId 			uint32
	appType 		uint32
	lastResponseTime	int64
}

type RouterLogic struct {
	mPool 			*PoolAndAgent.SingleMsgPool
	mListenAgent	*PoolAndAgent.NetAgent
	mMyAppId		uint32
	mMapConnection	map[uint64]uint32
	mMapAppId		map[uint32]RouterConnection
	mMapAppType		map[uint32]([]uint32)
	mRandNum		*rand.Rand
}

func (this *RouterLogic) Init(myPool *PoolAndAgent.SingleMsgPool) bool{
	this.mPool = myPool
	s2 := rand.NewSource(time.Now().UnixNano())
	this.mRandNum = rand.New(s2)
	this.mMapConnection = make(map[uint64]uint32)
	this.mMapAppId = make(map[uint32]RouterConnection)
	this.mMapAppType = make(map[uint32]([]uint32))
	return true
}

func (this *RouterLogic) ProcessReq(req proto.Message, pDatabase *PoolAndAgent.CADODatabase) {
	msg := Router_CreateCommonMsgByTCPTransferMsg(req)
	switch data := msg.(type) {
	case *PrivateInitMsg:
		this.Private_OnInit(data)
	case *tcpnet.TCPSessionCome:
		this.Network_OnConnOK(data)
	case *tcpnet.TCPSessionClose:
		this.Network_OnConnClose(data)
	case *router.RouterTransferData:
		fmt.Println("data_tra--->",data)
		this.Router_OnRouterDataReq(data)
	case *router.RegisterAppReq:
		fmt.Println("data_reg--->",data)
		this.Router_OnRegReq(data)
	default:
		return
	}

}

func (this *RouterLogic) OnPulse(ms uint64) {

}

func (this *RouterLogic) Private_OnInit(req *PrivateInitMsg) {
	this.mMyAppId = req.myAppId
}

func (this *RouterLogic) Router_OnRegReq(req *router.RegisterAppReq) {
	appId := req.AppId
	appType := req.AppType
	connId := req.Base.ConnId
	if _, ok1 := this.mMapConnection[connId];!ok1 {
		return
	}
	_, ok2 := this.mMapAppId[appId]
	if ok2{
		fmt.Println("相同类型相同app_id的app已经注册了，不允许重复注册。apptype=", appType, "typename=", bs_proto.GetAppTypeName(appType),
			",appid=", appId, ",connid=", req.Base.ConnId)
		log.Println("相同类型相同app_id的app已经注册了，不允许重复注册。apptype=", appType, "typename=", bs_proto.GetAppTypeName(appType),
			",appid=", appId, ",connid=", req.Base.ConnId)
		this.CloseSession(req.Base.ConnId)
		return
	}

	rsp := new(router.RegisterAppRsp)
	bs_proto.SetBaseKindAndSubId(rsp)
	bs_proto.CopyBaseExceptKindAndSubId(rsp.Base, req.Base)
	rsp.RegResult = 1
	this.SendToOtherApp(rsp, rsp.Base)

	this.mMapConnection[connId] = appId
	this.mMapAppId[appId] = RouterConnection{
		connId:           connId,
		isConnectted:     true,
		clientAdress:     req.Base.RemoteAdd,
		appId:            appId,
		appType:          appType,
		lastResponseTime: time.Now().Unix()}

	typeElem, ok3 := this.mMapAppType[appType]
	if !ok3 {
		sliceAppId := make([]uint32, 0, 10)
		this.mMapAppType[appType] = sliceAppId
		typeElem, ok3 = this.mMapAppType[appType]
	}
	var exist bool = false
	for _,v := range typeElem{
		if v == appId{
			exist = true
			break
		}
	}

	if exist{
		//对应的appId存在
		fmt.Println(fmt.Println("奇怪的错误，相同类型相同app_id的app在mMapAppType中找到了，",
			"但是却通过了前面的appid map和connid map的检查。apptype=", appType,
			",appid=", appId, ",connid=", req.Base.ConnId))
	}else{
		//对应的appId不存在,新增这个appId
		typeElem = append(typeElem, appId)
		this.mMapAppType[appType] = typeElem
		fmt.Println("mMapAppType=", this.mMapAppType)
	}
	fmt.Println("一个App注册来了,type=", appType,
		",typename=", bs_proto.GetAppTypeName(appType),
		",id=", appId)
}

func (this *RouterLogic) Network_OnConnOK(req *tcpnet.TCPSessionCome){
	if _, ok := this.mMapConnection[req.Base.ConnId]; ok {
		//FIXME 一般这不可能发生
		fmt.Println("发生了不可能事件，有重复的connId发生，connId = ", req.Base.ConnId)
	} else {
		fmt.Println("router 新建了一个客户端连接，connId = ", req.Base.ConnId)
		this.mMapConnection[req.Base.ConnId] = 0 //直到收到regreq后才知道app id，先设为0
	}
}
func (this *RouterLogic) Network_OnConnClose(req *tcpnet.TCPSessionClose){
	connId := req.Base.ConnId
	fmt.Println("conn_id=", req.Base.ConnId, "断开连接")
	appId, ok := this.mMapConnection[connId]
	if !ok{
		return
	}
	idElem, ok2 := this.mMapAppId[appId]
	if ok2 && idElem.connId == connId {
		if typeElem, ok3 := this.mMapAppType[idElem.appType];ok3{
			for i,v := range typeElem{
				if v== appId{
					typeElem = append(typeElem[:i], typeElem[i+1:]...)
					this.mMapAppType[idElem.appType] = typeElem
				}
			}
		}

		delete(this.mMapAppId, appId)
	}
	delete(this.mMapConnection, connId)
}

func (this *RouterLogic) Router_OnRouterDataReq(req *router.RouterTransferData) {
	connId := req.Base.ConnId
	destAppId := req.DestAppid
	destAppType := req.DestApptype
	if appId, ok1 := this.mMapConnection[connId]; !ok1 {
		fmt.Println("当前连接还没有注册，不转发其报文:connid=", connId)
		return
	}else {
		elem, ok2 := this.mMapAppId[appId]
		if ok2{
			elem.lastResponseTime = time.Now().Unix()
			this.mMapAppId[appId] = elem
			req.SrcApptype = elem.appType
			req.SrcAppid = elem.appId
		}
	}

	var sendResult bool = false
	switch destAppId {
	case uint32(mytype.EnumAppId_Send2AnyOne):
		sendResult = this.DeliverToAnyOneByType(req, destAppType)
	case uint32(mytype.EnumAppId_Send2All):
		sendResult = this.DeliverToAllByType(req, destAppType)
	default:
		sendResult = this.DeliverToAllByPointID(req, destAppType, destAppId)
	}
	if !sendResult{
		fmt.Println("destAppId:",destAppId)
		fmt.Println("目标APP无法找到，可能尚未注册",
			",dest apptype=", req.DestApptype,
			",dest appname=", bs_proto.GetAppTypeName(req.DestApptype),
			",dest appid=", req.DestAppid,
			",src apptype=", req.SrcApptype,
			",src appname=", bs_proto.GetAppTypeName(req.SrcApptype),
			",src appid=", req.SrcAppid,
			",src connid=", req.Base.ConnId,
			",cmd_kindid=", req.DataCmdKind,
			",cmd_subid=", req.DataCmdSubid,
			",userid=", req.AttUserid,
			",gate connid=", req.AttGateconnid)
	}else{
		fmt.Println("已向目标APP发送报文",
			",dest apptype=", req.DestApptype,
			",dest appname=", bs_proto.GetAppTypeName(req.DestApptype),
			",dest appid=", req.DestAppid,
			",src apptype=", req.SrcApptype,
			",src appname=", bs_proto.GetAppTypeName(req.SrcApptype),
			",src appid=", req.SrcAppid,
			",src connid=", req.Base.ConnId,
			",cmd_kindid=", req.DataCmdKind,
			",cmd_subid=", req.DataCmdSubid,
			",userid=", req.AttUserid,
			",gate connid=", req.AttGateconnid)
	}
}

//向客户端发送报文
func (this *RouterLogic) SendToOtherApp(req proto.Message, pBase *mytype.BaseInfo) {
	msg := Router_CreateTCPTransferMsgByCommonMsg(req, pBase)
	this.mPool.SendMsgToClientByNetAgent(msg)
}

func (this *RouterLogic) CloseSession(connId uint64) {
	kick := new(tcpnet.TCPSessionKick)
	bs_proto.SetBaseKindAndSubId(kick)
	kick.Base.ConnId = connId
	this.mPool.SendMsgToClientByNetAgent(kick)
}

func (this *RouterLogic) DeliverToAnyOneByType(req *router.RouterTransferData, destAppType uint32) bool{
	bs_proto.OutputMyLog("发送往任意的", bs_proto.GetAppTypeName(destAppType))
	if req.DataDirection == router.RouterTransferData_App2Client {
		bs_proto.OutputMyLog("服务端发往客户端的报文必须指定gate appid")
		return false
	}
	typeElem, ok2 := this.mMapAppType[destAppType]
	if !ok2{
		bs_proto.OutputMyLog("找不到相应的AppType=", bs_proto.GetAppTypeName(destAppType))
		return false
	}
	if size := len(typeElem); size > 0 {
		randIndex := this.mRandNum.Intn(size)
		randAppId := typeElem[randIndex]
		idElem, ok3 := this.mMapAppId[randAppId]
		if !ok3{
			bs_proto.OutputMyLog("找不到相应的AppId=", randAppId)
			return false
		}
		req.Base.ConnId = idElem.connId
		this.SendToOtherApp(req, req.Base)
	}else{
		return false
	}
	return true
}

func (this *RouterLogic) DeliverToAllByType(req *router.RouterTransferData, desAppType uint32) bool {
	if req.DataDirection == router.RouterTransferData_App2Client && desAppType == uint32((mytype.EnumAppType_Gate)) {
		//如果是服务端发往客户端的报文。接受发往所有gate，因为有可能是广播类型消息，但是要慎用，所以打印出来，以免滥用
		fmt.Println("此报文向将所有gate广播",
			",dest apptype=", req.DestApptype,
			",dest appname=", bs_proto.GetAppTypeName(req.DestApptype),
			",dest appid=", req.DestAppid,
			",src apptype=", req.SrcApptype,
			",src appname=", bs_proto.GetAppTypeName(req.SrcApptype),
			",src appid=", req.SrcAppid,
			",src connid=", req.Base.ConnId,
			",cmd_kindid=", req.DataCmdKind,
			",cmd_subid=", req.DataCmdSubid,
			",userid=", req.AttUserid,
			",gate connid=", req.AttGateconnid)
	}
	typeElem, ok2 := this.mMapAppType[desAppType]
	if !ok2 {
		return false
	}
	for _, appId := range typeElem {
		idElem, ok3 := this.mMapAppId[appId]
		if !ok3 {
			return false
		}
		req.Base.ConnId = idElem.connId
		this.SendToOtherApp(req, req.Base)
	}
	return true
}

func (this *RouterLogic) DeliverToAllByPointID(req *router.RouterTransferData, destAppType uint32, destAppId uint32) bool{
	if req.DataDirection == router.RouterTransferData_App2Client && destAppType != uint32(mytype.EnumAppType_Gate) {
		return false
	}
	idElem, ok3 := this.mMapAppId[destAppId]
	if !ok3 {
		return false
	}
	if idElem.appType != destAppType {
		return false
	}
	req.Base.ConnId = idElem.connId
	this.SendToOtherApp(req, req.Base)
	return true
}