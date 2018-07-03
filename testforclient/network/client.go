package network

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/uchihatmtkinu/RC/shard"
	"encoding/gob"
	"io/ioutil"
)

var nodeAddress string
var knownNodes = []string{"localhost:3000"}
var knownGroupNodes = []string{}
var myheight int

//version
type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

//address
type addr struct {
	AddrList []string
}

//command -> byte
func commandToBytes(command string) []byte {
	var bytees [commandLength]byte

	for i, c := range command {
		bytees[i] = byte(c)
	}

	return bytees[:]
}

//byte -> command
func bytesToCommand(bytees []byte) string {
	var command []byte

	for _, b := range bytees {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

//send data to addr
func sendData(addr string , data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

/*
func sendVersion(addr string, height int) {
	bestHeight := height
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})

	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}

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
		fmt.Print("update bestheigh to", myBestHeight)
		//sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, myheight)
	}

	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

//sen my addr to addr
func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(address, request)
}*/

// handle received address
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

//StartServer start a server
func StartServer(ID int) {



	ln, err := net.Listen(protocol, shard.MyMenShard.Address)
	fmt.Println("My IP+Port: ",shard.MyMenShard.Address)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	//TODO generate block.
	//bc := NewBlockchain(nodeID)

	//if account.MyAccount.Addr != knownNodes[0] {
	//	sendVersion(knownNodes[0], myheight)
	//}
	var command string
	var request []byte
	requestChannel := make(chan []byte, bufferSize)
	flag := true
	IntialReadyCh <- flag
	fmt.Println("intial ready")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, requestChannel)

		request = <-requestChannel
		if len(request) < commandLength {
			continue
		}
		command = bytesToCommand(request[:commandLength])
		if len(request) > commandLength {
			request = request[commandLength:]
		}
		fmt.Printf("Received %s command\n", command)
		// TODO instead of switch, we can use select to concurrently solve different commands
		switch command {

		case "Tx":
			go HandleAndSendTx(request)
		case "TxM":
			if shard.GlobalGroupMems[CacheDbRef.ID].Role == 0 {
				go HandleTxLeader(request)
			} else {
				go HandleTx(request)
			}
		case "TxList":
			go HandleTxList(request)
		case "TxDec":
			go HandleTxDecLeader(request)
		case "TxDecSet":
			go HandleAndSentTxDecSet(request)
		case "TxDecSetM":
			if shard.GlobalGroupMems[CacheDbRef.ID].Role == 0 {
				go HandleTxDecSetLeader(request)
			} else {
				go HandleTxDecSet(request)
			}
		case "TxB":
			go HandleTxBlock(request)
		case "FinalTxB":
			go HandleAndSentFinalTxBlock(request)
		case "FinalTxBM":
			go HandleFinalTxBlock(request)
		//shard
		case "shardReady":
			go HandleShardReady(request)
		case "readyAnnoun":
			go HandleShardReadyAnnounce(request)
		//rep pow
		case "RepPowAnnou":
			go HandleRepPowRx(request)

		//cosi protocol
		case "cosiAnnoun":
			go HandleCoSiAnnounce(request)

		case "cosiChallen":
			if CoSiFlag {
				go HandleCoSiChallenge(request)
			}
		case "cosiSig":
			if CoSiFlag {
				go HandleCoSiSig(request)
			}
		case "cosiCommit":
			if CoSiFlag {
				go HandleCoSiCommit(request)
			}
		case "cosiRespon":
			if CoSiFlag {
				go HandleCoSiResponse(request)
			}
		//sync
		case "requestSync":
			go HandleRequestSync(request)
		case "syncNReady":
			if SyncFlag {
				go HandleSyncNotReady(request)
			}
		case "syncSB":
			if SyncFlag {
				go HandleSyncSBMessage(request)
			}
		case "syncTB":
			if SyncFlag {
				go HandleSyncTBMessage(request)
			}
		default:
			fmt.Println("Unknown command!")
		}
	}
}

//encode
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

//whether it is a new node
func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
