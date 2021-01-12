package bs_proto

import (
	"fmt"
	"go_game/protodefine/client"
	"go_game/protodefine/gate"
	"go_game/protodefine/mytype"
	"go_game/protodefine/router"
	"go_game/protodefine/tcpnet"
	"reflect"
)

func setTcpBase(input interface{}) (bool, *mytype.BaseInfo) {
	fmt.Println("Data1111:",input)
	switch data := input.(type) {
	case *tcpnet.TCPTransferMsg:
		fmt.Println("报文类型为*bs_tcp.TCPTransferMsg")
		if data.Base == nil {
			fmt.Println("分配了new(mytype.BaseInfo)")
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindNetTCP)
		data.Base.SubId = uint32(tcpnet.CMDID_Tcp_IDTCPTransferMsg)
		//fmt.Println("Data:",data)
		return true, data.Base
	case *tcpnet.TCPSessionCome:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindNetTCP)
		data.Base.SubId = uint32(tcpnet.CMDID_Tcp_IDTCPSessionCome)
		//fmt.Println("Data:",data)
		return true, data.Base
	case *tcpnet.TCPSessionClose:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindNetTCP)
		data.Base.SubId = uint32(tcpnet.CMDID_Tcp_IDTCPSessionClose)
		//fmt.Println("Data:",data)
		return true, data.Base
	case *tcpnet.TCPSessionKick:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindNetTCP)
		data.Base.SubId = uint32(tcpnet.CMDID_Tcp_IDTCPSessionKick)
		//fmt.Println("Data:",data)
		return true, data.Base

	default:
		return false, nil
	}
}


func setGateBase(input interface{}) (bool, *mytype.BaseInfo) {
	switch data := input.(type) {
	case *gate.PulseReq:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindGate)
		data.Base.SubId = uint32(gate.CMDID_Gate_IDPulseReq)
		return true, data.Base
	case *gate.PulseRsp:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindGate)
		data.Base.SubId = uint32(gate.CMDID_Gate_IDPulseRsp)
		return true, data.Base
	case *gate.GateTransferData:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindGate)
		data.Base.SubId = uint32(gate.CMDID_Gate_IDTransferData)
		return true, data.Base
	default:
		return false, nil
	}
}

func setRouterBase(input interface{}) (bool, *mytype.BaseInfo) {
	switch data := input.(type) {
	case *router.RouterTransferData:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindRouter)
		data.Base.SubId = uint32(router.CMDID_Router_IDTransferData)
		return true, data.Base
	case *router.RegisterAppReq:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindRouter)
		data.Base.SubId = uint32(router.CMDID_Router_IDRegisterAppReq)
		return true, data.Base
	case *router.RegisterAppRsp:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindRouter)
		data.Base.SubId = uint32(router.CMDID_Router_IDRegisterAppRsp)
		return true, data.Base
	default:
		return false, nil
	}
}

func setClientBase(input interface{}) (bool, *mytype.BaseInfo) {
	switch data := input.(type) {
	case *client.LoginReq:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDLoginReq)
		return true, data.Base
	case *client.LoginRsp:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDLoginRsp)
		if data.UserBaseInfo == nil { //顺便设置一下复合数据类型
			data.UserBaseInfo = new(mytype.BaseUserInfo)
		}
		if data.UserSesionInfo == nil {
			data.UserSesionInfo = new(mytype.UserSessionInfo)
		}
		return true, data.Base
	case *client.LogoutReq:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDLogoutReq)
		return true, data.Base
	case *client.LogoutRsp:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDLogoutRsp)
		return true, data.Base
	case *client.QueryFundReq:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDQueryFundReq)
		return true, data.Base
	case *client.QueryFundRsp:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDQueryFundRsp)
		return true, data.Base
	case *client.GetOnlineUserReq:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDGetOnlineUserReq)
		return true, data.Base
	case *client.GetOnlineUserRsp:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDGetOnlineUserRsp)
		return true, data.Base
	case *client.KickUserReq:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDKickUserReq)
		return true, data.Base
	case *client.KickUserRsp:
		if data.Base == nil {
			data.Base = new(mytype.BaseInfo)
		}
		data.Base.KindId = uint32(mytype.CMDKindId_IDKindClient)
		data.Base.SubId = uint32(client.CMDID_Client_IDKickUserRsp)
		return true, data.Base
	default:
		return false, nil
	}
}

func  SetBaseKindAndSubId(input interface{}) (bool, *mytype.BaseInfo)  {
	if input ==nil{
		return false, nil
	}
	switch reflect.TypeOf(input).String(){
	case "*tcpnet.TCPTransferMsg":
		fallthrough
	case "*tcpnet.TCPSessionCome":
		fallthrough
	case "*tcpnet.TCPSessionClose":
		fallthrough
	case "*tcpnet.TCPSessionKick":
		return setTcpBase(input)
		//以下报文属于gate.proto,CMDKindId_IDKindGate大类
	case "*gate.PulseReq":
		fallthrough
	case "*gate.PulseRsp":
		fallthrough
	case "*gate.GateTransferData":
		return setGateBase(input)
	//以下报文属于router.proto,CMDKindId_IDKindRouter大类
	case "*router.RouterTransferData":
		fallthrough
	case "*router.RegisterAppReq":
		fallthrough
	case "*router.RegisterAppRsp":
		return setRouterBase(input)
	//以下报文属于client.proto,CMDKindId_IDKindClient大类
	case "*client.LoginReq":
		fallthrough
	case "*client.LoginRsp":
		fallthrough
	case "*client.LogoutReq":
		fallthrough
	case "*client.LogoutRsp":
		fallthrough
	case "*client.QueryFundReq":
		fallthrough
	case "*client.QueryFundRsp":
		fallthrough
	case "*client.GetOnlineUserReq":
		fallthrough
	case "*client.GetOnlineUserRsp":
		fallthrough
	case "*client.KickUserReq":
		fallthrough
	case "*client.KickUserRsp":
		return setClientBase(input)
	default:
	fmt.Println("input为不识别的类型")
	return false, nil
	}
	return false, nil
}

func CopyBaseExceptKindAndSubId(dst *mytype.BaseInfo, src *mytype.BaseInfo){
	if dst == nil || src == nil {
		fmt.Println("CopyBaseExceptKindAndSubId传入的参数是空,dst=", dst, ",src=", src)
		return
	}
	dst.ConnId = src.ConnId
	dst.GateConnId = src.GateConnId
	dst.RemoteAdd = src.RemoteAdd
	dst.AttApptype = src.AttApptype
	dst.AttAppid = src.AttAppid
}

func SetCommonMsgBaseByRouterTransferData(dst *mytype.BaseInfo, srcRouter *router.RouterTransferData){
	if dst == nil || srcRouter == nil {
		fmt.Println("SetCommonMsgBaseByRouterTransferData传入的参数是空,dst=", dst, ",src=", srcRouter)
		return
	}
	CopyBaseExceptKindAndSubId(dst, srcRouter.Base)
	//和客户端相关的都要重新赋值，取的是srcRouter的值，而不是srcRouter.Base的值
	dst.AttAppid = srcRouter.SrcAppid
	dst.AttApptype = srcRouter.SrcApptype
	dst.RemoteAdd = srcRouter.ClientRemoteAddress
	dst.GateConnId = srcRouter.AttGateconnid
}
