package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	bs_proto "go_game/protodefine"
	"go_game/protodefine/client"
	"go_game/protodefine/gate"
	"go_game/protodefine/mytype"
	"go_game/protodefine/router"
	"go_game/protodefine/tcpnet"
)

func Gate_CreateCommonMsgByTCPTransferMsg(req proto.Message) proto.Message {
	switch data := req.(type) {
	case *tcpnet.TCPTransferMsg:
		fmt.Println("bs_tcp.TCPTransferMsg base=", data.Base)
		switch data.DataKindId {
		case uint32(mytype.CMDKindId_IDKindGate):
			switch data.DataSubId {
			case uint32(gate.CMDID_Gate_IDPulseReq):
				msg := new(gate.PulseReq)
				err := proto.Unmarshal(data.Data, msg)
				bs_proto.SetBaseKindAndSubId(msg)
				bs_proto.CopyBaseExceptKindAndSubId(msg.Base, data.Base)
				if err != nil {
					fmt.Println("解析PulseReq出错")
					return nil
				}
				return msg
			case uint32(gate.CMDID_Gate_IDPulseRsp):
				msg := new(gate.PulseReq)
				err := proto.Unmarshal(data.Data, msg)
				bs_proto.SetBaseKindAndSubId(msg)
				bs_proto.CopyBaseExceptKindAndSubId(msg.Base, data.Base)
				if err != nil {
					fmt.Println("解析PulseReq出错")
					return nil
				}
				return msg
			case uint32(gate.CMDID_Gate_IDTransferData):
				msg := new(gate.GateTransferData)
				err := proto.Unmarshal(data.Data, msg)
				bs_proto.SetBaseKindAndSubId(msg)
				bs_proto.CopyBaseExceptKindAndSubId(msg.Base, data.Base)
				if err != nil {
					fmt.Println("解析TransferData出错")
					return nil
				}
				fmt.Println("成功解析TransferData")
				return msg
			default:
				fmt.Println("不识别的gate报文，DataSubId=", data.DataSubId)
				return nil //丢弃这个TCPTransferMsg报文
			}
		case uint32(mytype.CMDKindId_IDKindRouter):
			switch data.DataSubId { //判断小类
			case uint32(router.CMDID_Router_IDTransferData):
				msg := new(router.RouterTransferData)
				err := proto.Unmarshal(data.Data, msg)
				bs_proto.SetBaseKindAndSubId(msg)
				bs_proto.CopyBaseExceptKindAndSubId(msg.Base, data.Base)
				if err != nil {
					fmt.Println("解析RouterTransferData出错")
					return nil
				}
				return msg
			default:
				fmt.Println("不识别的router报文，DataSubId=", data.DataSubId)
				return nil //丢弃这个TCPTransferMsg报文
			}
		default:
			return nil
			}
	default:
		return req
	}
	return nil
}

func Gate_CreateTCPTransferMsgByCommonMsg(req proto.Message, pBase *mytype.BaseInfo) proto.Message {
	switch data := req.(type) {
	case *tcpnet.TCPSessionKick:
		return data
	case *tcpnet.TCPTransferMsg:
		return data
	default:
		msg := new(tcpnet.TCPTransferMsg)
		bs_proto.SetBaseKindAndSubId(msg)
		bs_proto.CopyBaseExceptKindAndSubId(msg.Base, pBase)
		msg.DataKindId = pBase.KindId
		msg.DataSubId = pBase.SubId
		buf, err := proto.Marshal(req)
		if err != nil {
			return nil
		} else {
			msg.Data = buf
			return msg
		}
	}
	return req
}

func Gate_CreateCommonMsgByRouterTransferMsg(req *router.RouterTransferData) proto.Message {
	if req == nil {
		return nil
	}

	//大类
	switch req.DataCmdKind {
	case uint32(mytype.CMDKindId_IDKindClient):
		//小类
		switch req.DataCmdSubid {
		case uint32(client.CMDID_Client_IDLoginRsp):
			msg := new(client.LoginRsp)
			err := proto.Unmarshal(req.Data, msg)
			if err != nil {
				fmt.Println("LoginRsp解析失败")
				return nil
			}
			bs_proto.SetBaseKindAndSubId(msg)
			bs_proto.SetCommonMsgBaseByRouterTransferData(msg.Base, req)
			return msg
		default:
			return nil
		}
	default:
		return nil
	}

	return nil
}

func Gate_CreateGateTransferMsgByCommonMsg(req proto.Message) *gate.GateTransferData {
	gateTrans := new(gate.GateTransferData)
	bs_proto.SetBaseKindAndSubId(gateTrans)

	switch data := req.(type) {
	case *router.RouterTransferData:
		//RouterTransferData要特殊对待，不是调用proto.Marshal
		bs_proto.CopyBaseExceptKindAndSubId(gateTrans.Base, data.Base)
		gateTrans.Base.ConnId = data.AttGateconnid
		gateTrans.DataCmdKind = data.DataCmdKind
		gateTrans.DataCmdSubid = data.DataCmdSubid
		gateTrans.Data = make([]byte, len(data.Data))
		copy(gateTrans.Data, data.Data) //使用copy，让req可以被GateConnection
		gateTrans.AttAppid = data.SrcAppid
		gateTrans.AttApptype = data.SrcApptype
		gateTrans.ReqId = 0
	case *client.LoginRsp:
		fmt.Println("序列化bs_client.LoginRsp, LoginRsp=", data)
		bs_proto.CopyBaseExceptKindAndSubId(gateTrans.Base, data.Base)
		buff, err := proto.Marshal(data)
		if err != nil {
			fmt.Println("LoginRsp序列化失败")
			return nil
		}
		gateTrans.Data = buff
		gateTrans.DataCmdKind = uint32(mytype.CMDKindId_IDKindClient)
		gateTrans.DataCmdSubid = uint32(client.CMDID_Client_IDLoginRsp)
		gateTrans.AttAppid = data.Base.AttAppid
		gateTrans.AttApptype = data.Base.AttApptype
		gateTrans.ClientRemoteAddress = data.Base.RemoteAdd
		fmt.Println("LoginRsp => gateTrans=", gateTrans)
	default:
		return nil
	}
	return gateTrans
}