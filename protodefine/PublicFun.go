package bs_proto

import (
	"fmt"
	"go_game/protodefine/mytype"
	"runtime"
)

func GetAppTypeName(appType uint32) string {
	var name string
	switch appType {
	case uint32(mytype.EnumAppType_Gate):
		name = "Gate"
	}
	return name
}

func OutputMyLog(a ...interface{}) {
	funcName, file, line, ok := runtime.Caller(1)
	if ok {
		fmt.Println("Func Name=", runtime.FuncForPC(funcName).Name())
		fmt.Printf("file: %s    line=%d\n", file, line)
		fmt.Println(a...)
	}
}
