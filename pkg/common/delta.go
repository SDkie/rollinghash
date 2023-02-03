package common

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/SDkie/rollinghash/pkg/rabinkarp"
)

type Delta struct {
	ChunkLen uint32
	Hashmap  map[uint32]uint32

	// -1 = no command
	// 0 = match
	// 1 = literal
	CurrCmd         int
	StartChunkIndex uint32
	EndChunkIndex   uint32
	Literals        []byte

	CurrChunk []byte
	Hash      uint32
	Pow       uint32

	OldFile   *os.File
	NewFile   *os.File
	DeltaFile *os.File
}

func NewDelta(sigFileName, oldFileName, newFileName, deltaFileName string) (*Delta, error) {
	sig, err := ReadSignature(sigFileName)
	if err != nil {
		return nil, err
	}

	var delta Delta
	delta.OldFile, err = os.Open(oldFileName)
	if err != nil {
		log.Printf("error opening oldFile: %s", err)
		return nil, err
	}
	delta.NewFile, err = os.Open(newFileName)
	if err != nil {
		log.Printf("Error opening newFile: %s", err)
		return nil, err
	}
	delta.DeltaFile, err = os.Create(deltaFileName)
	if err != nil {
		log.Printf("error creating deltaFile: %s", err)
		return nil, err
	}

	delta.CurrCmd = -1
	delta.ChunkLen = sig.ChunkLen
	delta.CurrChunk = make([]byte, delta.ChunkLen)
	delta.Hashmap = make(map[uint32]uint32)
	for i := uint32(0); i < sig.TotalChunks; i++ {
		delta.Hashmap[sig.Hashes[i]] = i
	}

	WriteUint32(delta.DeltaFile, delta.ChunkLen)
	if err != nil {
		return nil, err
	}

	return &delta, nil
}

func GenerateDelta(sigFileName, oldFileName, newFileName, deltaFileName string) error {
	delta, err := NewDelta(sigFileName, oldFileName, newFileName, deltaFileName)
	if err != nil {
		return err
	}
	defer delta.OldFile.Close()
	defer delta.DeltaFile.Close()
	defer delta.NewFile.Close()

	for {
		var err error
		if delta.CurrCmd == -1 || delta.CurrCmd == 0 {
			err = delta.ReadFullChunk()
		} else {
			err = delta.ReadNextByte()
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		index, ok := delta.Hashmap[delta.Hash]
		var match bool
		if ok {
			match, err = delta.CompareChunks(index)
			if err != nil {
				return err
			}
		}
		if match {
			err = delta.ChunkMatched(index)
		} else {
			err = delta.ChunkNotMatched()
		}
		if err != nil {
			return err
		}
	}
}

func (d *Delta) ReadFullChunk() error {
	_, err := d.NewFile.Read(d.CurrChunk)
	if err != nil {
		return err
	}
	d.Hash, d.Pow = rabinkarp.Hash(d.CurrChunk)
	return nil
}

func (d *Delta) ReadNextByte() error {
	b := make([]byte, 1)
	_, err := d.NewFile.Read(b)
	if err != nil {
		return err
	}

	d.Hash = rabinkarp.RollingHash(d.Hash, d.Pow, uint32(d.CurrChunk[0]), uint32(b[0]))
	d.CurrChunk = d.CurrChunk[1:]
	d.CurrChunk = append(d.CurrChunk, b[0])
	return nil
}

func (d *Delta) CompareChunks(chunkIndex uint32) (bool, error) {
	oldFileChunk := make([]byte, d.ChunkLen)

	_, err := d.OldFile.ReadAt(oldFileChunk, int64(chunkIndex*d.ChunkLen))
	if err != nil {
		log.Printf("Error reading file: %s", err)
		return false, err
	}

	return string(oldFileChunk) == string(d.CurrChunk), nil
}

func (d *Delta) ChunkMatched(index uint32) error {
	log.Printf("Chunk matched: %d\n", index)

	if d.CurrCmd == 0 && d.EndChunkIndex+1 == index {
		d.EndChunkIndex++
		d.Hash = 0
		d.Pow = 0
		return nil
	}

	if d.CurrCmd != -1 {
		err := d.WriteToDeltaFile()
		if err != nil {
			return err
		}
	}

	d.CurrCmd = 0
	d.StartChunkIndex = index
	d.EndChunkIndex = index
	d.Hash = 0
	d.Pow = 0
	return nil
}

func (d *Delta) ChunkNotMatched() error {
	log.Printf("Found literal: %s\n", string(d.CurrChunk[0]))

	if d.CurrCmd == 1 {
		d.Literals = append(d.Literals, d.CurrChunk[0])
		return nil
	}

	if d.CurrCmd == 0 {
		err := d.WriteToDeltaFile()
		if err != nil {
			return err
		}
	}

	d.CurrCmd = 1
	d.Literals = d.CurrChunk[:1]
	return nil
}

func (d *Delta) WriteToDeltaFile() error {
	var content string
	if d.CurrCmd == 0 {
		content = fmt.Sprintf("%02x%03x%03x", d.CurrCmd, d.StartChunkIndex, d.EndChunkIndex)

	} else {
		content = fmt.Sprintf("%02x%06x", d.CurrCmd, len(d.Literals))
		content += string(d.Literals)
	}

	data, err := hex.DecodeString(content)
	if err != nil {
		log.Printf("Error decoding hex string: %s", err)
		return err
	}
	_, err = d.DeltaFile.Write(data)
	if err != nil {
		log.Printf("Error writing to delta file: %s", err)
		return err
	}

	return nil
}
