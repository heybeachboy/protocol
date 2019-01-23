package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

/**
 *@author MR.zhou
 *@mail zhouletian1234@live.com
 *@comment implement ICMP protocol with using golang
 *@date 2019-01-23 13:53:54
 */
const TIME_OUT = 20 * time.Second //返回最长超时请求

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
func (i *ICMP) Ping() {

}

/**
 *创建ICMP数据结构
 */

func (i *ICMP) CreateICMP(sep uint16) (*ICMP) {
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
	return &icmp

}

/**
 *发送ICMP数据包
 */
func (i *ICMP) SendICMPPacket(icmp ICMP, address *net.IPAddr) (error) {
	conn, err := net.DialIP("ip4:icmp", nil, address)
	if err != nil {
		fmt.Printf("Fail to connect to remote host: %s\n", err)
		return err
	}
	defer conn.Close()
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

	fmt.Printf("%d bytes from %s: seq=%d time=%dms\n", n, address.String(), icmp.SequenceNum, duration)

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

func main() {

}
