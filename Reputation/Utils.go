package Reputation

import (
	"bytes"
	"encoding/binary"
	"log"
)



// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// BoolToHex converts a bool to byte array
func BoolToHex(f bool) []byte {
	var a []byte
	if f {
		a = append(a,1)
	} else {
		a = append(a,0)
	}

	return a
}


/*
func UIntToHex(num uint64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}*/