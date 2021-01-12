package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

var quit chan bool = make(chan bool)

func main()  {
	//var tcpAddr *net.TCPAddr
	//tcpAddr, _ = net.ResolveTCPAddr("tcp", "127.0.0.1:9999")
	//tcpListener, _ := net.ListenTCP("tcp", tcpAddr)
	//defer tcpListener.Close()
	//
	//for {
	//	tcpConn,err := tcpListener.AcceptTCP()
	//	if err!= nil{
	//		continue
	//	}
	//
	//	fmt.Println("A client connected: "+tcpConn.RemoteAddr().String())
	//	go tcpPipe(tcpConn)
	//}

}

func tcpPipe(tcpConn *net.TCPConn) {
	ipStr := tcpConn.RemoteAddr().String()
	defer func() {
		fmt.Println("disconnected: "+ ipStr)
		tcpConn.Close()
	}()

	reader := bufio.NewReader(tcpConn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		fmt.Println(string(message))
		msg := time.Now().String() + "\n"
		b := []byte(msg)
		tcpConn.Write(b)
	}
}
