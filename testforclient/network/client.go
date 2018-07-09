package network

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"encoding/gob"
	"io/ioutil"

	"github.com/uchihatmtkinu/RC/shard"
)

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
func sendData(addr string, data []byte) {
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
	fmt.Println("My IP+Port: ", shard.MyMenShard.Address)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

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

		request := <-requestChannel
		if len(request) < commandLength {
			continue
		}
		command := bytesToCommand(request[:commandLength])
		if len(request) > commandLength {
			request = request[commandLength:]
		}
		//fmt.Printf("%d Received %s command\n", ID, command)
		switch command {
		case "requestTxB":
			go HandleRequestTxB(request)
		case "Tx":
			go HandleAndSendTx(request)
		case "TxM":
			go HandleTotalTx(request)
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
			go HandleFinalTxBlock(request)
		case "StartTxB":
			go HandleStartTxBlock(request)
		//shard
		case "shardReady":
			go HandleShardReady(request)
		case "readyAnnoun":
			go HandleShardReady(request)
		case "leaderReady":
			go HandleLeaderReady(request)
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
