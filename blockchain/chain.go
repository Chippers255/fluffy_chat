package blockchain

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Node struct{}

type Transaction struct {
	Sender    string  `json:"sender"`
	Amount    float32 `json:"amount"`
	Recipient string  `json:"recipient"`
}

type Message struct {
	Sender string `json:"sender"`
	Text   string `json:"text"`
	Chat   string `json:"chat"`
}

type Block struct {
	Index        int           `json:"index"`
	Proof        string        `json:"proof"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PreviousHash string        `json:"previous_hash"`
}

type Blockchain struct {
	Chain               []Block         `json:"chain"`
	CurrentTransactions []Transaction   `json:"current_transactions"`
	Nodes               map[string]Node `json:"nodes"`
}

type Server interface {
	RegisterNode(address string)
	ValidChain(chain []Block) bool
	ResolveConflicts()
	NewBlock() Block
	NewTransaction(sender string, receiver string, amount float32) int
	LastBlock() Block
	Hash(block Block) string
	WorkProof(last Block) string
	ValidProof(lastProof string, proof string, lastHash string) bool
}

func (b *Blockchain) RegisterNode(address string) {
	parsedUrl, _ := url.Parse(address)

	b.Nodes[parsedUrl.Host] = Node{}
}

func (b *Blockchain) ValidChain(chain []Block) bool {
	lastBlock := chain[0]
	currentIndex := 1

	for currentIndex < len(chain) {
		block := chain[currentIndex]
		fmt.Println(lastBlock)
		fmt.Println(block)
		fmt.Println("-----------------------------------------------------------")

		if block.PreviousHash != Hash(lastBlock) {
			return false
		}

		if !ValidProof(lastBlock.Proof, block.Proof) {
			return false
		}

		lastBlock = block
		currentIndex += 1
	}

	return true
}

func (b *Blockchain) ResolveConflicts() bool {
	neighbours := b.Nodes
	var newChain []Block

	maxLength := len(b.Chain)

	for _, node := range neighbours {
		response, _ := http.Get(fmt.Sprintf("http://%s/chain", node))

		if response.StatusCode == 200 {
			var respy resp2
			_ = json.NewDecoder(response.Body).Decode(&respy)

			length := respy.Length
			chain := respy.Chain

			if length > maxLength && b.ValidChain(chain) {
				maxLength = length
				newChain = chain
			}
		}
	}

	if newChain != nil {
		b.Chain = newChain
		return true
	}

	return false
}

func (b *Blockchain) LastBlock() Block {
	return b.Chain[len(b.Chain)-1]
}

func (b *Blockchain) NewTransaction(newTransaction Transaction) int {
	b.CurrentTransactions = append(b.CurrentTransactions, newTransaction)

	return b.LastBlock().Index + 1
}

func (b *Blockchain) NewBlock(proof string, previousHash string) Block {
	block := Block{
		Index:        len(b.Chain) + 1,
		Proof:        proof,
		Timestamp:    time.Now().Unix(),
		Transactions: b.CurrentTransactions,
		PreviousHash: previousHash,
	}

	b.CurrentTransactions = nil
	b.Chain = append(b.Chain, block)

	return block
}

func (b *Blockchain) ProofOfWork(lastProof string) string {
	proof := 0

	for !ValidProof(lastProof, string(proof)) {
		proof += 1
	}

	return fmt.Sprintf("%d", proof)
}

type Resp struct {
	Message string `json:"message"`
}

func ValidProof(lastProof string, proof string) bool {
	guess := fmt.Sprintf("%s%s", lastProof, proof)
	guessHash := md5.Sum([]byte(guess))
	guessHex := fmt.Sprintf("%x", guessHash)

	if string(guessHex[0]) == "0" {
		return true
	} else {
		return false
	}
}

func Hash(block Block) string {
	s, _ := json.Marshal(block)
	s2 := md5.Sum(s)

	d := fmt.Sprintf("%x", s2)

	return d
}

func NewBlockchain() Blockchain {
	blockchain := Blockchain{
		Chain:               []Block{},
		CurrentTransactions: []Transaction{},
		Nodes:               map[string]Node{},
	}

	_ = blockchain.NewBlock("100", "1")

	return blockchain
}