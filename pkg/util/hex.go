package util

import (
	"encoding/binary"
	"log"
	"os"
)

// WriteUint32InHex converts decimal uint32 number into hex and writes to given file
func WriteUint32InHex(file *os.File, n uint32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	_, err := file.Write(b)
	if err != nil {
		log.Printf("error writing to file: %s", err)
		return err
	}
	return nil
}
