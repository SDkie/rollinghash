package signature

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/SDkie/rollinghash/pkg/rabinkarp"
	"github.com/SDkie/rollinghash/pkg/util"
)

// Signature File Format:
// 4 bytes - chunk length
// 4 bytes - hash for each chunk

// GenerateSignature generates a signature file for a given input file.
func GenerateSignature(inputFileName, sigFileName string) error {
	infile, err := os.Open(inputFileName)
	if err != nil {
		log.Printf("error opening input file: %s", err)
		return err
	}
	defer infile.Close()

	sigfile, err := os.OpenFile(sigFileName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("error creating signature file: %s", err)
		return err
	}
	defer sigfile.Close()

	stats, err := infile.Stat()
	if err != nil {
		log.Printf("error getting file stats: %s", err)
		return err
	}

	fileSize := stats.Size()
	if fileSize == 0 {
		err := fmt.Errorf("input filesize is 0")
		log.Println(err)
		return err
	}

	chunkSize := getOptimalChunkSize(fileSize)
	err = util.WriteUint32InHex(sigfile, uint32(chunkSize))
	if err != nil {
		return err
	}

	log.Printf("File size: %d", fileSize)
	log.Printf("Chunk size: %d", chunkSize)

	chunk := make([]byte, chunkSize)
	var n int
	for i := 0; ; i++ {
		n, err = infile.Read(chunk)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("error reading file: %s", err)
			return err
		}
		chunk = chunk[:n]

		hash, _ := rabinkarp.Hash(chunk)
		err = util.WriteUint32InHex(sigfile, hash)
		if err != nil {
			return err
		}
		log.Printf("Chunk %d: Hash: %08x", i, hash)
	}

	return nil
}

// Signature contains all the information stored in a signature file
type Signature struct {
	ChunkLen    uint32
	TotalChunks uint32
	Hashes      []uint32
}

// ReadSignature reads a signature file and returns a Signature struct.
func ReadSignature(sigFileName string) (*Signature, error) {
	sigfile, err := os.Open(sigFileName)
	if err != nil {
		log.Printf("error opening signature File: %s", err)
		return nil, err
	}
	defer sigfile.Close()

	stats, err := sigfile.Stat()
	if err != nil {
		log.Printf("error getting file stats: %s", err)
		return nil, err
	}
	if stats.Size()%4 != 0 {
		err := fmt.Errorf("invalid signature file")
		log.Println(err)
		return nil, err
	}

	var signature Signature
	data := make([]byte, 4)

	_, err = sigfile.Read(data)
	if err != nil {
		log.Printf("error reading signature file: %s", err)
		return nil, err
	}

	signature.ChunkLen = binary.BigEndian.Uint32(data)
	if signature.ChunkLen < 256 || signature.ChunkLen%128 != 0 {
		err := fmt.Errorf("invalid chunk size")
		log.Println(err)
		return nil, err
	}

	signature.TotalChunks = uint32(stats.Size()/4) - 1
	log.Printf("ChunkLen: %d", signature.ChunkLen)
	log.Printf("TotalChunks: %d", signature.TotalChunks)

	for i := uint32(0); i < signature.TotalChunks; i++ {
		_, err = sigfile.Read(data)
		if err != nil {
			log.Printf("error reading file: %s", err)
			return nil, err
		}
		hash := binary.BigEndian.Uint32(data)
		signature.Hashes = append(signature.Hashes, hash)
		log.Printf("Chunk %d: Hash: %08x", i, hash)
	}

	return &signature, nil
}
