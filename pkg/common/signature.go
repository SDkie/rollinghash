package common

import (
	"io"
	"log"
	"os"

	"github.com/SDkie/rollinghash/pkg/rabinkarp"
)

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
