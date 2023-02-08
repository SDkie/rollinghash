# rollinghash

---
* [Intro](#intro)
* [Build](#build)
* [Run](#run)
* [Testing](#testing)
---
## Intro:
- `rollinghash` allows to create signature and delta of given file. 
- `signature` sub-command creates signature of input-file
- `delta` sub-command creates delta-file which can be used to convert original-file to updated-file
- `delta` sub-command needs signature and original file both, as just matching of hash can't guarantee matching of the chunks

## Build
    go build ./cmd/rollinghash

## Run
Create signature file:

    ./rollinghash signature <input_file> <signature_file>
    
Create delta file:

    ./rollinghash delta <original_file> <signature_file> <updated_file> <delta_file>

## Testing
    go test ./...
