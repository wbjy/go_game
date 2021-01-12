package DialManager

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	ListenManager "go_game/TcpManager"
	"go_game/protodefine/tcpnet"
	"net"
)

//作为客户端去连接

type ConnectSession struct {
	remoterAddr string // 对端地址包括ip和port
	connId uint64		// 连接标识id
	tcpConn  *net.TCPConn	//用于发送接收消息的tcp连接实体
	cache	*bytes.Buffer	// 自动扩展的用于战报处理的缓存区
	MsgWriteCh	chan *tcpnet.TCPTransferMsg	//从接收来自逻辑层的消息管道
	Quit	chan bool
}

func (session *ConnectSession) RecvPackage(logicChannel chan proto.Message) {
	data := make([]byte, 1024)
	for {
		dataLenght,err := (*session).tcpConn.Read(data)
		if err != nil {
			fmt.Println("TCP读取数据错误:", err.Error())
			kick := new(tcpnet.TCPTransferMsg)
			session.tcpConn.Close()
			session.MsgWriteCh <- kick
			session.Quit <- true	// 通知主线程断开
			break
		}else {
			buff := data[:dataLenght]
			validBuff := ListenManager.DecodePackage(session.cache, buff)
			if validBuff == nil {
				continue  //报文没读完整继续read
			}
			protoMsg := new(tcpnet.TCPTransferMsg)
			err := proto.Unmarshal(validBuff, protoMsg)
			if err != nil {
				fmt.Println("反序列化TCPTransferMsg报文出错，无法解析validBuff, err=", err.Error())
				//可能消息出错了清空cache
				session.cache.Reset()
			}else {
				protoMsg.Base.ConnId = session.connId
				logicChannel <- protoMsg
			}
		}
	}
}

// 向客户端发消息
func (session *ConnectSession) SendPackage(logicChannel chan proto.Message) {
	quit := false
	for {
		if quit {
			break
		}

		select {
		case v, ok := <- session.MsgWriteCh:
			if ok {
				fmt.Println(v)
				data, err := proto.Marshal(v)
				if err != nil {
					fmt.Println("序列化TCPTransferMsg报文出错")
				} else {
					var pkg *bytes.Buffer = new(bytes.Buffer)
					ListenManager.EncodePackage(pkg, data)
					_, err2 := session.tcpConn.Write(pkg.Bytes())
					if err2 != nil {
						//直接tcpConn.close就行了，因为RecvPackege协程是阻塞在read处的，这里close了，read就能得到error异常了
						fmt.Println("TCP写数据错误:", data, err2.Error())
						session.tcpConn.Close()
						session.Quit <- true //通知主线程连接断开了
						quit = true
					}
				}
			}else {
				//表示关闭了通道。session也需要关闭
				quit = true
				fmt.Println("connId=", session.connId, "的会话已经退出send协程")
			}
		}
	}
}

func CreateClien(remote_address string, logicChannel chan proto.Message) *ConnectSession {
	conn, err := net.Dial("tcp", remote_address)
	if err != nil {
		fmt.Println("连接服务端失败:", err.Error(), remote_address)
		return nil
	}
	var tcpConn *net.TCPConn
	switch myconn := conn.(type) {
	case *net.TCPConn:
		fmt.Println("已建立起连接的服务器TCP连接，对端地址=", myconn.RemoteAddr().String())
		tcpConn = myconn
	default:
		return nil
	}
	fmt.Println("DialManager connected : ", tcpConn.RemoteAddr().String())
	session := new(ConnectSession)
	session.remoterAddr = tcpConn.RemoteAddr().String()
	session.tcpConn = tcpConn
	session.cache = new(bytes.Buffer)
	session.MsgWriteCh = make(chan *tcpnet.TCPTransferMsg, 10000)
	session.Quit = make(chan bool)
	go session.RecvPackage(logicChannel)
	go session.SendPackage(logicChannel)

	return session
}