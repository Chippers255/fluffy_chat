package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"net/http"
	"strings"
)

type resp struct {
	Message      string        `json:"message"`
	Index        int           `json:"index"`
	Transactions []Transaction `json:"transactions"`
	Proof        string        `json:"proof"`
	PreviousHash string        `json:"previous_hash"`
}

type resp2 struct {
	Chain  []Block `json:"chain"`
	Length int     `json:"length"`
}

func (b *Blockchain) mine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	lastBlock := b.LastBlock()
	lastProof := lastBlock.Proof
	proof := b.ProofOfWork(lastProof)

	nt := Transaction{
		Sender:    "0",
		Amount:    1,
		Recipient: viper.GetString("NODE_ID"),
	}

	b.NewTransaction(nt)
	previousHash := Hash(lastBlock)
	block := b.NewBlock(proof, previousHash)

	response := resp{
		Message:      "New Block Forged",
		Index:        block.Index,
		Transactions: block.Transactions,
		Proof:        block.Proof,
		PreviousHash: block.PreviousHash,
	}

	_ = json.NewEncoder(w).Encode(&response)
}

func (b *Blockchain) chain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := resp2{
		Chain:  b.Chain,
		Length: len(b.Chain),
	}

	_ = json.NewEncoder(w).Encode(&response)
}

func (b *Blockchain) newTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var transaction Transaction

	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		fmt.Println(err)
	}
	index := b.NewTransaction(transaction)

	response := Resp{
		Message: fmt.Sprintf("Transaction will be added to Block %d", index),
	}

	_ = json.NewEncoder(w).Encode(&response)
}

type registerRequest struct {
	Nodes []string `json:"nodes"`
}

type registerResponse struct {
	Message string `json:"message"`
	TotalNodes map[string]Node `json:"total_nodes"`
}

func (b *Blockchain) register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request registerRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println(err)
	}

	nodes := request.Nodes
	// TODO: Handle and empty request brother
	if nodes == nil {
		fmt.Println("Error: Please supply a valid list of nodes")
	}

	for _, node := range nodes {
		b.RegisterNode(node)
	}

	response := registerResponse{
		Message:    "New nodes have been added",
		TotalNodes: b.Nodes,
	}

	_ = json.NewEncoder(w).Encode(&response)
}

type consensusResponse struct {
	Message string `json:"message"`
	Chain []Block `json:"chain"`
}

func (b *Blockchain) consensus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	replaced := b.ResolveConflicts()

	if replaced {
		response := consensusResponse{
			Message: "Our chain was replaced",
			Chain:   b.Chain,
		}

		_ = json.NewEncoder(w).Encode(&response)
	} else {
		response := consensusResponse{
			Message: "Our chain is authoritative",
			Chain:   b.Chain,
		}

		_ = json.NewEncoder(w).Encode(&response)
	}
}

func main() {
	nodeID := strings.ReplaceAll(uuid.New().String(), "-", "")
	viper.Set("NODE_ID", nodeID)

	router := mux.NewRouter()
	b := NewBlockchain()

	// Dispatch map for CRUD operations.
	router.HandleFunc("/mine", b.mine).Methods("GET")
	router.HandleFunc("/chain", b.chain).Methods("GET")
	router.HandleFunc("/transactions/new", b.newTransaction).Methods("POST")
	router.HandleFunc("/nodes/register", b.register).Methods("POST")
	router.HandleFunc("/nodes/resolve", b.consensus).Methods("Get")

	// Start the server.
	port := ":5000"
	fmt.Println("\nListening on port " + port)
	_ = http.ListenAndServe(port, router)
}