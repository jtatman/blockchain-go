package main


// imports
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)


// create structure of block on chain
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}


// blockchain
var Blockchain []Block

// block function to create hash of blockdata
func calculateHash(block Block) string {
    record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
    h := sha256.New()
    h.Write([]byte(record))
    hashed := h.Sum(nil)
    return hex.EncodeToString(hashed)
}

// generate block function
func generateBlock(oldBlock Block, BPM int)(Block, error) {
    var newBlock Block

    t := time.Now()
    newBlock.Index = oldBlock.Index + 1
    newBlock.Timestamp = t.String()
    newBlock.BPM = BPM
    newBlock.PrevHash = oldBlock.Hash
    newBlock.Hash = calculateHash(newBlock)

    return newBlock, nil
}

// function to check validity of blocks on chain
func isBlockValid( newBlock, oldBlock Block) bool {
    if oldBlock.Index+1 != newBlock.Index {
        return false
    }
    if oldBlock.Hash != newBlock.PrevHash {
        return false
    }
    if calculateHash(newBlock) != newBlock.Hash {
        return false
    }
    return true

}

// check if blockchain is valid from a length measurement
// longest chain wins when sharing
func replaceChain(newBlocks []Block) {
    if len(newBlocks) > len(Blockchain) {
        Blockchain = newBlocks
    }

}

// initialize gorilla mux router
// initialize handlers
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

// uses gorilla:mux to create webserver
// uses port from .env file
func run() error {
	mux := makeMuxRouter()
	httpPort := os.Getenv("PORT")
    httpAddr := os.Getenv("ADDR")
	// log.Println("Listening on ", os.Getenv("PORT"))
    log.Println("Listening on ", httpAddr,  ":", httpPort)
	s := &http.Server{
		Addr:           "192.168.1.144:" + httpPort,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// return blockchain in JSON format
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

// structure for bpm measurement recording
type Message struct {
    BPM int
}

// http handler for formatting
 
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

// error handler that prints back json when needed
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

// main function to tie it back together
func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal(err)
    }

    go func() {
        t := time.Now()
        genesisBlock := Block{0, t.String(), 0, "", ""}
        spew.Dump(genesisBlock)
        Blockchain = append(Blockchain, genesisBlock)
    }()
    log.Fatal(run())

}

