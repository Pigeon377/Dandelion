package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

}

const (
	port = 11451
)

func startClient() {
	conn, err := net.Dial("tcp", "0.0.0.0:11451")
	if err != nil {
		fmt.Printf("dial failed, err:%v\n", err)
		return
	}

	fmt.Println("Conn Established...:")

	reader := bufio.NewReader(os.Stdin)
	for {
		data, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("read from console failed, err:%v\n", err)
			break
		}
		data = strings.TrimSpace(data)
		_, err = conn.Write([]byte(data))
		if err != nil {
			fmt.Printf("write failed, err:%v\n", err)
			break
		}
	}
}
