package network

import (
	"fmt"
	"net"
	"encoding/gob"
	"io/ioutil"
	"log"
	"bytes"
	"io"
)



var nodeAddress string
var knownNodes = []string{"localhost:3000"}
var knownGroupNodes = []string{}
var myheight int


type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}



type addr struct {
	AddrList []string
}



func commandToBytes(command string) []byte {
	var bytees [commandLength]byte

	for i, c := range command {
		bytees[i] = byte(c)
	}

	return bytees[:]
}

func bytesToCommand(bytees []byte) string {
	var command []byte

	for _, b := range bytees {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}


//TODO handle data in block
//func handleGetData(request []byte, bc *Blockchain)

//TODO sync
func sendVersion(addr string, height int) {
	bestHeight := height
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})

	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}
//TODO sync
func handleVersion(request []byte, height int) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := height
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		myBestHeight = foreignerBestHeight
		fmt.Print("update bestheigh to",myBestHeight)
		//sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, myheight)
	}

	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}





func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(address, request)
}

// add
func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	//requestBlocks()
}



// handle connection
func handleConnection(conn net.Conn, requestChannel chan []byte) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	requestChannel <- request

}


func StartServer(nodeID string, height int) {
	//assume in one computer.
	//nodeAddress = fmt.Sprintf("%s", nodeID)
	myheight = height
	MyAccount.New(nodeID)
	ln, err := net.Listen(protocol, MyAccount.Addr)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	//TODO generate block.
	//bc := NewBlockchain(nodeID)

	if MyAccount.Addr != knownNodes[0] {
		sendVersion(knownNodes[0], myheight)
	}
	var command string
	var request []byte
	requestChannel := make(chan []byte, bufferSize)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, requestChannel)

		request = <- requestChannel
		command = bytesToCommand(request[:commandLength])
		fmt.Printf("Received %s command\n", command)
		// TODO instead of switch, we can use select to concurrently solve different commands
		switch command{
		case "add":
			handleAddr(request)
		case "version":
			handleVersion(request, myheight)
		default:
			fmt.Println("Unknown command!")
		}
	}
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}



func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}