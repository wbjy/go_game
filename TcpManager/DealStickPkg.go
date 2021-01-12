package ListenManager

import (
	"bytes"
	"encoding/binary"
)

func DecodePackage(pkg *bytes.Buffer, recvData []byte) []byte {
	//log.Println(recvData, len(recvData))
	pkg.Write(recvData)
	if pkg.Len() < 4 {
		return nil
	}
	var length int32 = 0
	tmppkg := new(bytes.Buffer)
	tmppkg.Write(pkg.Bytes()[:4])
	binary.Read(tmppkg, binary.BigEndian, &length)
	if int(length)+4 > pkg.Len() {
		return nil
	}else {
		binary.Read(pkg, binary.BigEndian, &length)
		buff := make([]byte, length)
		binary.Read(pkg, binary.BigEndian, buff)
		return buff
	}
	return nil
}

//发送消息组包
// 这个和erlang的打包规则差不多，也是先发长度，再发列表内容
func EncodePackage(pkg *bytes.Buffer, data []byte) error {
	var length int32 = int32(len(data))
	err := binary.Write(pkg, binary.BigEndian, length)
	if err != nil {
		return err
	}
	err = binary.Write(pkg, binary.BigEndian, data)
	if err!= nil {
		return err
	}
	return nil
}
