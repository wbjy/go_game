package ListenManager

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	bs_proto "go_game/protodefine"
	"go_game/protodefine/tcpnet"
	"net"
	"sync"
	"sync/atomic"
)

var gSessionMap map[uint64]*ConnectionSession
var gSessionLock sync.Mutex

//服务端
type ConnectionSession struct {
	remoteAddr string   		//对端地址ip+port
	connId 	   uint64			//连接标识id
	tcpConn		*net.TCPConn	//用于发送接收消息tcp连接实体
	cache		*bytes.Buffer   //自动扩展的用于粘包处理的缓存区
	MsgWriteCh  chan *tcpnet.TCPTransferMsg //从接受来自逻辑层消息的管道
	wg 			sync.WaitGroup		//在close时阻塞等待两个协程结束
	isWriteClosed 	int32
	isReadClosed 	int32
	IsSendKickMsg 	int32
}


func CreateServer(ip_address string, logicChannel chan proto.Message){
	var tcpAddr *net.TCPAddr
	tcpAddr,_ = net.ResolveTCPAddr("tcp", ip_address)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("监听端口失败：",err.Error())
		return
	}
	fmt.Println("已初始化连接，等待客户端连接...", ip_address)

	var currentConnId uint64 = 0
	gSessionMap = make(map[uint64]*ConnectionSession)
	for{
		tcpConn, err := listener.AcceptTCP()
		if err != nil{
			fmt.Println("accepct错误:", err.Error())
			continue
		}
		currentConnId += 1
		fmt.Println("A client connected : ", tcpConn.RemoteAddr().String(), "currentConnId=", currentConnId)
		session := new(ConnectionSession)
		session.remoteAddr = tcpConn.RemoteAddr().String()
		session.connId = currentConnId
		session.tcpConn = tcpConn
		session.cache = new(bytes.Buffer)
		session.MsgWriteCh = make(chan *tcpnet.TCPTransferMsg, 100)
		session.isReadClosed = 0
		session.isWriteClosed = 0
		session.IsSendKickMsg = 0
		gSessionLock.Lock()
		gSessionMap[currentConnId] = session
		gSessionLock.Unlock()
		session.wg.Add(2)
		go session.RecvPackage(logicChannel)
		go session.SendPackage(logicChannel)
		msg := new(tcpnet.TCPSessionCome)
		bs_proto.SetBaseKindAndSubId(msg)
		msg.Base.ConnId = session.connId
		msg.Base.RemoteAdd = session.remoteAddr
		logicChannel <- msg
		fmt.Println("向logicChannel发送了TCPSessionCome报文")
	}

}
//接包
func (session *ConnectionSession) RecvPackage(logicChannel chan proto.Message) {
	data := make([]byte, 1024)
	for {
		dataLength,err := (*session).tcpConn.Read(data)
		if err != nil {
			fmt.Println("read client error:",err.Error())
			kick := new(tcpnet.TCPSessionKick)
			bs_proto.SetBaseKindAndSubId(kick)
			kick.Base.ConnId = session.connId
			logicChannel <- kick
			break
		}else {
			buff := data[:dataLength]
			fmt.Println(session.remoteAddr, session.cache)
			validBuff := DecodePackage(session.cache, buff)
			if validBuff == nil{
				continue
			}
			protoMsg := new(tcpnet.TCPTransferMsg)
			err := proto.Unmarshal(validBuff, protoMsg)
			if err != nil {
				fmt.Println("反序列化TCPTransferMsg报文出错，无法解析validBuff, err=", err.Error())
				//可能消息出错了清空cache
				session.cache.Reset()
			}else {
				protoMsg.Base.ConnId = session.connId
				protoMsg.Base.GateConnId = session.connId
				protoMsg.Base.RemoteAdd = session.remoteAddr
				logicChannel <- protoMsg
			}
		}
	}
	atomic.StoreInt32(&(session.isReadClosed), 1)
	session.wg.Done()
}

func (session *ConnectionSession) SendPackage(logicChannel chan proto.Message) {
	quit := false
	for {
		if quit {
			break
		}

		select {
		case v, ok := <- session.MsgWriteCh:
			if ok {
				data, err := proto.Marshal(v)
				if err != nil {
					fmt.Println("序列化TCPTransferMsg报文出错")
				}else {
					var pkg *bytes.Buffer = new(bytes.Buffer)
					EncodePackage(pkg, data)
					_, err2 := session.tcpConn.Write(pkg.Bytes())
					if err2 != nil {
						kick := new(tcpnet.TCPSessionKick)
						bs_proto.SetBaseKindAndSubId(kick)
						kick.Base.ConnId = session.connId
						logicChannel <- kick
						quit = true
					}
				}
			}else {
				quit = true
				fmt.Println("connId=", session.connId, "的会话已经退出send协程")
			}
		}
	}
	atomic.StoreInt32(&(session.isReadClosed), 1)
	session.wg.Done()
}

func (session *ConnectionSession) CloseSession(logicChannel chan proto.Message){
	ret := atomic.CompareAndSwapInt32(&(session.IsSendKickMsg), 0, 1)
	if ! ret {
		fmt.Println("已经CloseSession被调用了两次，connId=", session.connId)
	}
	if v := atomic.LoadInt32(&(session.isWriteClosed)); v == 0{
		close(session.MsgWriteCh) //关闭SendPackage接收消息的管道，以结束SendPackage协程
	}
	session.wg.Wait() //阻塞CloseSesion直到读写两个协程都结束
	connId := session.connId
	gSessionLock.Lock()
	delete(gSessionMap, session.connId)
	gSessionLock.Unlock()
	msg := new(tcpnet.TCPSessionClose)
	bs_proto.SetBaseKindAndSubId(msg)
	msg.Base.ConnId = connId
	logicChannel <- msg
	fmt.Println("gSessionMap已删除connId=", connId, "的session")
}


func GetSessionByConnId(connId uint64) *ConnectionSession {
	if elem, ok := gSessionMap[connId]; ok {
		return elem
	}
	return nil
}