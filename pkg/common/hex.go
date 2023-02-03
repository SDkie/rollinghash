package common

import (
	"encoding/binary"
	"log"
	"os"
)

func WriteUint32(file *os.File, n uint32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	_, err := file.Write(b)
	if err != nil {
		log.Printf("error writing int to file: %s", err)
		return err
	}
	return nil
}
