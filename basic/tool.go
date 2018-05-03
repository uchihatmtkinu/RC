package basic

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
)

//EncodeByte add length bit at the head of the byte array
func EncodeByte(current *[]byte, d *[]byte) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint32(len(*d)))
	if err != nil {
		return fmt.Errorf("Basic.EncodeByte write failed, %s", err)
	}
	tmp := append(buf.Bytes(), *d...)
	*current = append(*current, tmp...)
	return nil
}

//DecodeByte Decode bytes with a uint32 preamble
func DecodeByte(d *[]byte, out *[]byte) error {
	tmpBit := binary.LittleEndian.Uint32((*d)[:4])
	if 4+tmpBit > uint32(len(*d)) {
		return fmt.Errorf("Basic.DecodeByte length not enough")
	}
	tmpArray := (*d)[4 : 4+tmpBit]
	*d = (*d)[4+tmpBit:]
	*out = tmpArray
	return nil
}

//EncodeByteL add length bit at the head of the byte array with specific length
func EncodeByteL(current *[]byte, d *[]byte, l int) error {
	if len(*d) != l {
		return fmt.Errorf("Basic.EncodeByteL length is not right")
	}
	*current = append(*current, *d...)
	return nil
}

//DecodeByteL Decode bytes with a uint32 preamble
func DecodeByteL(d *[]byte, out *[]byte, l int) error {
	if len(*d) < l {
		return fmt.Errorf("Basic.DecodeByteL length not enough")
	}
	*out = (*d)[:l]
	*d = (*d)[l:]
	return nil
}

//EncodeInt Encode the data
func EncodeInt(current *[]byte, d interface{}) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, d)
	if err != nil {
		return fmt.Errorf("Basic.EncodeInt write failed, %s", err)
	}
	*current = append(*current, buf.Bytes()...)
	return nil
}

//DecodeInt Decode the data to uint32
func DecodeInt(d *[]byte, out interface{}) error {
	var bitlen uint32

	switch out := out.(type) {
	case *uint32:
		bitlen = 4
		tmpBit := binary.LittleEndian.Uint32((*d)[:bitlen])
		if bitlen+tmpBit > uint32(len(*d)) {
			return fmt.Errorf("Basic.DecodeInt length not enough")
		}
		*out = tmpBit
	case *uint64:
		bitlen = 8
		tmpBit := binary.LittleEndian.Uint64((*d)[:bitlen])
		if bitlen+uint32(tmpBit) > uint32(len(*d)) {
			return fmt.Errorf("Basic.DecodeInt length not enough")
		}
		*out = tmpBit
	default:
		bitlen = 0
	}

	*d = (*d)[bitlen:]

	return nil
}

//EncodeBig encodes a big number into byte[]
func EncodeBig(current *[]byte, d *big.Int) error {
	tmp := uint32(len(d.Bytes()))
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, tmp)
	if err != nil {
		return fmt.Errorf("Basic.EncodeBig write failed, %s", err)
	}
	*current = append(*current, buf.Bytes()...)
	*current = append(*current, d.Bytes()...)
	return nil
}

//DecodeBig decodes a byte[] into a big number
func DecodeBig(current *[]byte, out *big.Int) error {
	var bitlen uint32
	err := DecodeInt(current, &bitlen)
	if err != nil {
		return err
	}
	if uint32(len(*current)) < bitlen {
		return fmt.Errorf("Basic.DecodeBig not enough length")
	}
	out.SetBytes((*current)[:bitlen][:])
	fmt.Println(*out)
	*current = (*current)[bitlen:]
	return nil
}

//EncodeDoubleBig encodes a signature into data
func EncodeDoubleBig(current *[]byte, a *big.Int, b *big.Int) error {
	tmp := *current
	err := EncodeBig(&tmp, a)
	if err != nil {
		return err
	}
	err = EncodeBig(&tmp, b)
	if err != nil {
		return err
	}
	*current = tmp
	return nil
}

//DecodeDoubleBig encodes a signature into data
func DecodeDoubleBig(current *[]byte, a *big.Int, b *big.Int) error {
	err := DecodeBig(current, a)
	if err != nil {
		return err
	}
	err = DecodeBig(current, b)
	if err != nil {
		return err
	}
	return nil
}
