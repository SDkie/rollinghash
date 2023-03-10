package signature

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"

	"github.com/SDkie/rollinghash/pkg/rabinkarp"
	"github.com/SDkie/rollinghash/pkg/util"
)

// Signature File Format:
// 4 bytes - chunk length
// 4 bytes - hash for each chunk

var (
	ErrEmptyInputFile       = errors.New("inputFile is empty")
	ErrInvalidSignatureFile = errors.New("invalid signature file")
	ErrInvalidChunkSize     = errors.New("invalid chunk size")
)

// Signature contains all the information stored in a signature file
type Signature struct {
	ChunkLen    uint32
	TotalChunks uint32
	Hashes      []uint32
}

// GenerateSignature generates a signature file for a given input file.
func GenerateSignature(inputFileName, sigFileName string) (*Signature, error) {
	var signature Signature

	// Input file
	infile, err := os.Open(inputFileName)
	if err != nil {
		log.Printf("error opening input file: %s", err)
		return nil, err
	}
	defer infile.Close()
	stats, err := infile.Stat()
	if err != nil {
		log.Printf("error getting file stats: %s", err)
		return nil, err
	}
	fileSize := stats.Size()
	if fileSize == 0 {
		err := ErrEmptyInputFile
		log.Println(err)
		return nil, err
	}

	signature.ChunkLen = getOptimalChunkSize(fileSize)

	log.Printf("File size: %d", fileSize)
	log.Printf("Chunk size: %d", signature.ChunkLen)

	chunk := make([]byte, signature.ChunkLen)
	for i := 0; ; i++ {
		n, err := infile.Read(chunk)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("error reading file: %s", err)
			return nil, err
		}
		chunk = chunk[:n]

		hash, _ := rabinkarp.Hash(chunk)
		signature.Hashes = append(signature.Hashes, hash)
		log.Printf("Chunk %d: Hash: %08x", i, hash)
	}

	err = signature.write(sigFileName)
	return &signature, err
}

// write creates the signature file and writes the signature to it
func (s *Signature) write(filename string) error {
	sigfile, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("error creating signature file: %s", err)
		return err
	}
	defer sigfile.Close()

	err = util.WriteUint32InHex(sigfile, s.ChunkLen)
	if err != nil {
		return err
	}

	for _, hash := range s.Hashes {
		err = util.WriteUint32InHex(sigfile, hash)
		if err != nil {
			return err
		}
	}

	return nil
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
	if stats.Size() < 8 || stats.Size()%4 != 0 {
		err := ErrInvalidSignatureFile
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
		err := ErrInvalidChunkSize
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
