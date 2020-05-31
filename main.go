package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type Subscriber struct {
	ip   string
	recv chan string
}

var subs = make(map[string]*Subscriber)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: ", os.Args[0], "host")
		os.Exit(1)
	}

	host := os.Args[1]

	agent := os.Args[2]
	fmt.Println("Starting agent: " + agent)

	id := os.Args[3]

	switch agent {
	case "server":
		runServer(host)
	case "pub":
		runPub(host, id)
	case "sub":
		runSub(host, id)
	default:
		fmt.Println("bad agent")
		os.Exit(1)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}

func handleSubConnection(conn *net.TCPConn, c chan string) {
	for {
		defer func() {
			fmt.Println("closing connection")
			conn.Close()
		}()

		send := <-c

		_, err := conn.Write([]byte(send))
		if err != nil {

		}
	}
}

func handlePubConnection(conn *net.TCPConn) {
	msg := [512]byte{}
	for {
		defer func() {
			fmt.Println("closing connection")
			conn.Close()
		}()

		_, err := conn.Read(msg[0:])
		checkError(err)

		for _, s := range subs {
			s.recv <- string(msg[0:])
		}
		msg = [512]byte{}
	}
}

func runServer(host string) error {
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		fmt.Println("Resolution error", err.Error())
		os.Exit(1)
	}

	l, err := net.ListenTCP("tcp", addr)
	checkError(err)

	defer l.Close()

	var msg [512]byte

	for {
		conn, err := l.AcceptTCP()
		checkError(err)

		// Read first message
		_, err = conn.Read(msg[0:])
		checkError(err)

		client := string(msg[0:3])

		switch client {
		case "PUB":
			fmt.Println("NEW PUBLISTER: " + conn.RemoteAddr().String())
			go handlePubConnection(conn)
		case "SUB":
			fmt.Println("NEW SUBSCRIBER: " + conn.RemoteAddr().String())
			recv := make(chan string)
			subs[conn.RemoteAddr().String()] = &Subscriber{
				ip:   conn.RemoteAddr().String(),
				recv: recv,
			}
			go handleSubConnection(conn, recv)
		default:
			fmt.Println("not a valid client " + client)
			os.Exit(1)
		}
	}
}

func runSub(host, id string) error {
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		fmt.Println("Resolution error", err.Error())
		os.Exit(1)
	}

	var msg [512]byte
	conn, err := net.DialTCP("tcp", nil, addr)
	checkError(err)

	_, err = conn.Write([]byte("SUB"))
	checkError(err)

	for {
		defer func() {
			fmt.Println("closing connection")
			conn.Close()
		}()
		_, err := conn.Read(msg[0:])
		checkError(err)

		fmt.Println(string(msg[0:]))
	}
}

func runPub(host, id string) error {
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		fmt.Println("Resolution error", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	checkError(err)

	_, err = conn.Write([]byte("PUB"))
	checkError(err)

	for {
		defer func() {
			fmt.Println("closing connection")
			conn.Close()
		}()
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		checkError(err)
		_, err = conn.Write([]byte(strings.Trim(text, "\n")))
		checkError(err)
	}
}
