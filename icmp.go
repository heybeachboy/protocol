package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

/**
 *@author MR.zhou
 *@mail zhouletian1234@live.com
 *@comment implement ICMP protocol with using golang
 *@date 2019-01-23 13:53:54
 */
const TIME_OUT = 1 * time.Second //返回最长超时请求
var (
	sigChan         = make(chan os.Signal)
	QUIT_FLAG       int32 = 0
	conn            net.Conn
	ipString        string
	err             error
)

type ICMP struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	Identifier  uint16
	SequenceNum uint16
}

/**
 *ping 程序
 */
func (i *ICMP) Ping(host string) {
	i.initConnection(host)
	count := 0
	for {
		quit := atomic.LoadInt32(&QUIT_FLAG)
		if quit != 0 {
			break
		}

		i.SendICMPPacket(i.CreateICMP(uint16(count)))
		time.Sleep(500 * time.Millisecond)
		count++
	}

}

func (i *ICMP) initConnection(host string) {
	addr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		fmt.Printf("Fail to resolve %s, %s\n", host, err)
		return
	}

	conn, err = net.DialIP("ip4:icmp", nil, addr)
	if err != nil {
		fmt.Printf("Fail to connect to remote host: %s\n", err)
		os.Exit(1)
	}else {
		ipString = addr.String()
        fmt.Printf("Connection to host : %s successfuly\n",ipString)
	}


}

/**
 *创建ICMP数据结构
 */

func (i *ICMP) CreateICMP(sep uint16) (ICMP) {
	var icmp ICMP
	icmp.Type = 8
	icmp.Code = 0
	icmp.CheckSum = 0
	icmp.Identifier = 0
	icmp.SequenceNum = sep
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.CheckSum = i.checkSum(buffer.Bytes())
	buffer.Reset()
	return icmp

}

/**
 *发送ICMP数据包
 */
func (i *ICMP) SendICMPPacket(icmp ICMP) (error) {

	//defer conn.Close()
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		return err
	}
	start := time.Now()
	conn.SetReadDeadline(start.Add(TIME_OUT))
	reply := make([]byte, 1024)
	n, err := conn.Read(reply)
	reply = reply[:n]

	if err != nil {
		return err
	}

	end := time.Now()
	duration := end.Sub(start).Nanoseconds() / 1e6

	fmt.Printf("%d bytes from %s: seq=%d time=%dms\n", n, ipString, icmp.SequenceNum, duration)
	return err
}

/**
 *数据完整性校验
 */
func (i *ICMP) checkSum(data []byte) (uint16) {
	var (
		sum    uint32
		length = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)

	return uint16(^sum)

}

/**
 *捕获系统信号处理
 */
func catchSystemSignal() {
	for sig := range sigChan {
		switch sig {
		case syscall.SIGQUIT, syscall.SIGINT: //重新信号处理
			atomic.StoreInt32(&QUIT_FLAG, 1)
			conn.Close()
			os.Exit(0)
		default:
			fmt.Println("signal : ", sig)

		}

	}

}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("host is null: %s", os.Args[0])
		os.Exit(1)
	}
	signal.Notify(sigChan, syscall.SIGILL, syscall.SIGQUIT, syscall.SIGINT)
	i := ICMP{}
	go i.Ping(os.Args[1])
	catchSystemSignal() //主要处理退出处理

}
