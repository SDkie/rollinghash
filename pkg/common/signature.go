package common

import (
	"encoding/binary"
	"io"
	"log"
	"os"

	"github.com/SDkie/rollinghash/pkg/rabinkarp"
)

// GenerateSignature generates a signature file for a given input file.
func GenerateSignature(inputFileName, outputFileName string) error {
	infile, err := os.Open(inputFileName)
	if err != nil {
		log.Printf("Error opening InputFile: %s", err)
		return err
	}
	defer infile.Close()

	outfile, err := os.Create(outputFileName)
	if err != nil {
		log.Printf("Error creating SignatureFile: %s", err)
		return err
	}
	defer outfile.Close()

	stats, err := infile.Stat()
	if err != nil {
		log.Printf("Error getting file stats: %s", err)
		return err
	}

	fileSize := stats.Size()
	chunkSize := getOptimalChunkSize(fileSize)

	WriteUint32(outfile, uint32(chunkSize))
	if err != nil {
		return err
	}

	log.Printf("File size: %d", stats.Size())
	log.Printf("Chunk size: %d", chunkSize)

	chunk := make([]byte, chunkSize)

	for i := 0; ; i++ {
		_, err = infile.Read(chunk)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading file: %s", err)
			return err
		}

		hash, _ := rabinkarp.Hash(chunk)
		err = WriteUint32(outfile, hash)
		if err != nil {
			return err
		}
		log.Printf("Chunk %d: Hash: %08x", i, hash)
	}

	return nil
}

type Signature struct {
	ChunkLen    uint32
	TotalChunks uint32
	Hashes      []uint32
}

// ReadSignature reads a signature file and returns a Signature struct.
func ReadSignature(sigFileName string) (*Signature, error) {
	infile, err := os.Open(sigFileName)
	if err != nil {
		log.Printf("Error opening SignatureFile: %s", err)
		return nil, err
	}
	defer infile.Close()

	var signature Signature
	data := make([]byte, 4)

	_, err = infile.Read(data)
	if err != nil {
		log.Printf("Error reading file: %s", err)
		return nil, err
	}
	signature.ChunkLen = binary.BigEndian.Uint32(data)

	log.Printf("ChunkLen: %d", signature.ChunkLen)

	for i := 0; ; i++ {
		_, err = infile.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading file: %s", err)
			return nil, err
		}
		hash := binary.BigEndian.Uint32(data)
		signature.Hashes = append(signature.Hashes, hash)
		signature.TotalChunks++
		log.Printf("Chunk %d: Hash: %08x", i, hash)
	}

	return &signature, nil
}
