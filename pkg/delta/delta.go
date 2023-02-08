package delta

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/SDkie/rollinghash/pkg/rabinkarp"
	"github.com/SDkie/rollinghash/pkg/signature"
	"github.com/SDkie/rollinghash/pkg/util"
)

// CmdType is used for creating delta file
// 00 in the delta file means match
// 01 in the delta file means miss (literal)
type CmdType int

const (
	NO_CMD CmdType = iota - 1
	MATCH
	LITERAL
)

// Delta struct contains all the data required to generate delta file
type Delta struct {
	chunkLen uint32
	hashmap  map[uint32]uint32

	currCmd         CmdType
	startChunkIndex uint32
	endChunkIndex   uint32
	literals        []byte

	currChunk []byte
	hash      uint32
	pow       uint32

	oldFile   *os.File
	newFile   *os.File
	deltaFile *os.File
}

// newDelta create a new Delta struct
// it opens all the provided files
// also reads the signature file and insert all the hashes in a hashmap
func newDelta(oldFileName, sigFileName, newFileName, deltaFileName string) (*Delta, error) {
	sig, err := signature.ReadSignature(sigFileName)
	if err != nil {
		return nil, err
	}

	var delta Delta
	delta.oldFile, err = os.Open(oldFileName)
	if err != nil {
		log.Printf("error opening oldFile: %s", err)
		return nil, err
	}
	delta.newFile, err = os.Open(newFileName)
	if err != nil {
		log.Printf("error opening newFile: %s", err)
		return nil, err
	}
	delta.deltaFile, err = os.OpenFile(deltaFileName, os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		log.Printf("error creating deltaFile: %s", err)
		return nil, err
	}

	delta.currCmd = NO_CMD
	delta.chunkLen = sig.ChunkLen
	delta.currChunk = make([]byte, delta.chunkLen)
	delta.hashmap = make(map[uint32]uint32)
	for i := uint32(0); i < sig.TotalChunks; i++ {
		delta.hashmap[sig.Hashes[i]] = i
	}

	util.WriteUint32InHex(delta.deltaFile, delta.chunkLen)
	if err != nil {
		return nil, err
	}

	return &delta, nil
}

// GenerateDelta generates the delta file
func GenerateDelta(oldFileName, sigFileName, newFileName, deltaFileName string) error {
	delta, err := newDelta(oldFileName, sigFileName, newFileName, deltaFileName)
	if err != nil {
		return err
	}
	defer delta.oldFile.Close()
	defer delta.deltaFile.Close()
	defer delta.newFile.Close()

	for {
		var err error
		if delta.currCmd == NO_CMD || delta.currCmd == MATCH {
			err = delta.readFullChunk()
		} else {
			err = delta.readNextByte()
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// got EOF if len(delta.CurrChunk) < int(delta.ChunkLen)
		if len(delta.currChunk) < int(delta.chunkLen) {
			break
		}

		log.Printf("searching Hash: %08x", delta.hash)

		index, ok := delta.hashmap[delta.hash]
		var match bool
		if ok {
			match, err = delta.compareChunks(index)
			if err != nil {
				return err
			}
		}
		if match {
			err = delta.chunkFound(index)
		} else {
			err = delta.literalFound()
		}
		if err != nil {
			return err
		}
	}

	for len(delta.currChunk) > 0 {
		index, ok := delta.hashmap[delta.hash]
		var match bool
		if ok {
			match, err = delta.compareChunks(index)
			if err != nil {
				return err
			}
		}
		if match {
			err = delta.chunkFound(index)
			delta.currChunk = []byte{}
		} else {
			err = delta.literalFound()
			delta.skipFirstByte()
		}
		if err != nil {
			return err
		}
	}

	return delta.writeToDeltaFile()
}

// readFullChunk tries to read the fullChunk from the newFile
func (d *Delta) readFullChunk() error {
	n, err := d.newFile.Read(d.currChunk)
	if err != nil {
		if err == io.EOF {
			d.currChunk = []byte{}
			d.hash = 0
			d.pow = 0
		}
		return err
	}
	d.currChunk = d.currChunk[:n]
	d.hash, d.pow = rabinkarp.Hash(d.currChunk)
	return nil
}

// readNextByte tries to read the next byte and rotate the chunk
func (d *Delta) readNextByte() error {
	b := make([]byte, 1)
	_, err := d.newFile.Read(b)
	if err != nil {
		return err
	}

	d.hash = rabinkarp.Rotate(d.hash, d.pow, uint32(d.currChunk[0]), uint32(b[0]))
	d.currChunk = d.currChunk[1:]
	d.currChunk = append(d.currChunk, b[0])
	return nil
}

// skipFirstByte skips the first byte of the currChunk and calculates the hash
func (d *Delta) skipFirstByte() {
	d.hash = rabinkarp.Rotate(d.hash, d.pow, uint32(d.currChunk[0]), 0)
	d.currChunk = d.currChunk[1:]
}

// compareChunks compares the currChunk with the chunk from the oldFile
func (d *Delta) compareChunks(chunkIndex uint32) (bool, error) {
	oldFileChunk := make([]byte, d.chunkLen)

	_, err := d.oldFile.ReadAt(oldFileChunk, int64(chunkIndex*d.chunkLen))
	if err != nil {
		log.Printf("error reading file: %s", err)
		return false, err
	}

	return string(oldFileChunk) == string(d.currChunk), nil
}

// chunkFound is called when currChunk matches with a chunk in oldFile
func (d *Delta) chunkFound(index uint32) error {
	log.Printf("Chunk matched: %d\n", index)

	if d.currCmd == MATCH && d.endChunkIndex+1 == index {
		d.endChunkIndex++
		d.hash = 0
		d.pow = 0
		return nil
	}

	if d.currCmd != NO_CMD {
		err := d.writeToDeltaFile()
		if err != nil {
			return err
		}
	}

	d.currCmd = MATCH
	d.startChunkIndex = index
	d.endChunkIndex = index
	d.hash = 0
	d.pow = 0
	return nil
}

// literalFound is called when literal is found
// that is because currChunk does not match with any chunk in oldFile
func (d *Delta) literalFound() error {
	log.Printf("Found literal: %s\n", string(d.currChunk[0]))

	if d.currCmd == LITERAL {
		d.literals = append(d.literals, d.currChunk[0])
		return nil
	}

	if d.currCmd == MATCH {
		err := d.writeToDeltaFile()
		if err != nil {
			return err
		}
	}

	d.currCmd = LITERAL
	d.literals = d.currChunk[:1]
	return nil
}

// writeToDeltaFile writes the current command to the delta file
func (d *Delta) writeToDeltaFile() error {
	var content string
	if d.currCmd == MATCH {
		content = fmt.Sprintf("%02x%03x%03x", d.currCmd, d.startChunkIndex, d.endChunkIndex)
	} else if d.currCmd == LITERAL {
		content = fmt.Sprintf("%02x%06x", d.currCmd, len(d.literals))
	} else {
		err := fmt.Errorf("can't write invalid command:%d to delta file", d.currCmd)
		log.Println(err)
		return err
	}

	data, err := hex.DecodeString(content)
	if err != nil {
		log.Printf("error decoding hex string: %s", err)
		return err
	}
	_, err = d.deltaFile.Write(data)
	if err != nil {
		log.Printf("error writing to delta file: %s", err)
		return err
	}

	if d.currCmd == LITERAL {
		_, err = d.deltaFile.Write(d.literals)
		if err != nil {
			log.Printf("error writing literals to delta file: %s", err)
			return err
		}
		d.literals = nil
	}

	return nil
}
