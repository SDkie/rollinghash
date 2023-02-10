package delta

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/SDkie/rollinghash/pkg/rabinkarp"
	"github.com/SDkie/rollinghash/pkg/signature"
	"github.com/SDkie/rollinghash/pkg/util"
)

var (
	ErrEmptyOriginalFile = errors.New("originalFile is empty")
	ErrEmptyUpdatedFile  = errors.New("updatedFile is empty")
)

// Delta File Format:
// 4 bytes - chunk length
// 4 bytes - cmd with chunk index or literal with size
// if chunk match:
// 	    '00'      - cmd
//      'XXX'     - start chunk index
//	    'XXX'     - end chunk index
// if literal:
//	    '01'      - cmd (1 byte)
//      'XXXXXX'  - literal size (3 bytes)
// in case of literal after the cmd and size, literal data is written

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
type delta struct {
	chunkLen uint32
	hashmap  map[uint32]uint32

	currCmd         CmdType
	startChunkIndex uint32
	endChunkIndex   uint32
	literals        []byte

	currChunk []byte
	hash      uint32
	pow       uint32

	originalFile *os.File
	updatedFile  *os.File
	deltaFile    *os.File
}

// newDelta create a new Delta struct
// it opens all the provided files
// also reads the signature file and insert all the hashes in a hashmap
func newDelta(originalFile, sigFile, updatedFile, deltaFile string) (*delta, error) {
	var d delta
	// Signature file
	sig, err := signature.ReadSignature(sigFile)
	if err != nil {
		return nil, err
	}
	d.chunkLen = sig.ChunkLen
	d.hashmap = make(map[uint32]uint32)
	for i := uint32(0); i < sig.TotalChunks; i++ {
		d.hashmap[sig.Hashes[i]] = i
	}

	//  Old file
	d.originalFile, err = os.Open(originalFile)
	if err != nil {
		log.Printf("error opening originalFile: %s", err)
		return nil, err
	}
	stats, err := d.originalFile.Stat()
	if err != nil {
		log.Printf("error getting originalFile stats: %s", err)
		return nil, err
	}
	if stats.Size() == 0 {
		err := ErrEmptyOriginalFile
		log.Println(err)
		return nil, err
	}

	// New file
	d.updatedFile, err = os.Open(updatedFile)
	if err != nil {
		log.Printf("error opening updatedFile: %s", err)
		return nil, err
	}
	stats, err = d.updatedFile.Stat()
	if err != nil {
		log.Printf("error getting updatedFile stats: %s", err)
		return nil, err
	}
	if stats.Size() == 0 {
		err := ErrEmptyUpdatedFile
		log.Println(err)
		return nil, err
	}

	// Delta file
	d.deltaFile, err = os.OpenFile(deltaFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("error creating deltaFile: %s", err)
		return nil, err
	}

	d.currCmd = NO_CMD
	d.currChunk = make([]byte, d.chunkLen)

	err = util.WriteUint32InHex(d.deltaFile, d.chunkLen)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (d *delta) cleanup() {
	d.originalFile.Close()
	d.deltaFile.Close()
	d.updatedFile.Close()
}

// GenerateDelta generates the delta file
// signature and original file both are required for genearing delta,
// as just matching of hash can't guarantee matching of the chunks
func GenerateDelta(oldFileName, sigFileName, newFileName, deltaFileName string) error {
	d, err := newDelta(oldFileName, sigFileName, newFileName, deltaFileName)
	if err != nil {
		return err
	}
	defer d.cleanup()

	for {
		if d.currCmd == NO_CMD || d.currCmd == MATCH {
			err = d.readFullChunk()
		} else {
			err = d.readNextByte()
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// got EOF if len(d.CurrChunk) < int(d.ChunkLen)
		if len(d.currChunk) < int(d.chunkLen) {
			break
		}

		err = d.searchAndUpdate()
		if err != nil {
			return err
		}
	}

	for len(d.currChunk) > 0 {
		err = d.searchAndUpdate()
		if err != nil {
			return err
		}

		if d.currCmd == LITERAL {
			d.skipFirstByte()
		} else {
			d.currChunk = []byte{}
		}
	}

	return d.writeToDeltaFile()
}

// readFullChunk tries to read the fullChunk from the newFile
func (d *delta) readFullChunk() error {
	n, err := d.updatedFile.Read(d.currChunk)
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
func (d *delta) readNextByte() error {
	b := make([]byte, 1)
	_, err := d.updatedFile.Read(b)
	if err != nil {
		if err == io.EOF {
			d.skipFirstByte()
			return nil
		}
		return err
	}

	d.hash = rabinkarp.Rotate(d.hash, d.pow, uint32(d.currChunk[0]), uint32(b[0]))
	d.currChunk = d.currChunk[1:]
	d.currChunk = append(d.currChunk, b[0])
	return nil
}

// skipFirstByte skips the first byte of the currChunk and calculates the hash
func (d *delta) skipFirstByte() {
	d.hash, d.pow = rabinkarp.RollOut(d.hash, d.pow, uint32(d.currChunk[0]))
	d.currChunk = d.currChunk[1:]
}

// update searching chunk in oldFile and updates the delta based on that
func (d *delta) searchAndUpdate() error {
	match, index, err := d.searchChunk()
	if err != nil {
		return err
	}

	if match {
		return d.chunkFound(index)
	}
	return d.literalFound()
}

// searchChunk searches for the currChunk in oldFile
func (d *delta) searchChunk() (bool, uint32, error) {
	log.Printf("searching Hash: %08x", d.hash)
	index, ok := d.hashmap[d.hash]
	if !ok {
		return false, 0, nil
	}

	//read the chunk from oldFile and compare the content
	oldFileChunk := make([]byte, d.chunkLen)
	n, err := d.originalFile.ReadAt(oldFileChunk, int64(index*d.chunkLen))
	if err != nil && err != io.EOF {
		log.Printf("error reading file: %s", err)
		return false, 0, err
	}
	oldFileChunk = oldFileChunk[:n]

	if string(oldFileChunk) != string(d.currChunk) {
		return false, 0, nil
	}

	return true, index, nil
}

// chunkFound is called when currChunk matches with a chunk in oldFile
func (d *delta) chunkFound(index uint32) error {
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
func (d *delta) literalFound() error {
	log.Printf("Found literal: %s\n", string(d.currChunk[0]))

	if d.currCmd == MATCH {
		err := d.writeToDeltaFile()
		if err != nil {
			return err
		}
	}

	d.currCmd = LITERAL
	d.literals = append(d.literals, d.currChunk[0])
	return nil
}

// writeToDeltaFile writes the current command to the delta file
func (d *delta) writeToDeltaFile() error {
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
