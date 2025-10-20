package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Block struct {
	Index       int
	Timestamp   int64
	Transaction []Transaction
	PrevHash    string
	Proof       int
	Hash        string
}

type Transaction struct {
	Sender    string
	HashID    string
	Recipient string
	Amount    float64
}

type Blockchain struct {
	Chain        []Block
	Transactions []Transaction
}

func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		Chain:        []Block{},
		Transactions: []Transaction{},
	}
	bc.createGenesisBlock()
	return bc
}

func HashGenerate(block Block) string {
	hashInput := fmt.Sprintf("%d%d%s%d%d", block.Index, block.Timestamp, block.PrevHash, block.Proof, len(block.Transaction))
	for _, tx := range block.Transaction {
		hashInput += tx.Sender + tx.Recipient + fmt.Sprintf("%f", tx.Amount)
	}
	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}

func HashID(tx Transaction) string {
	data := fmt.Sprintf("%s%s%f", tx.Sender, tx.Recipient, tx.Amount)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])

}

func (bk *Blockchain) AddBlock(proof int, timestamp int64, prevHash string) {
	newBlock := Block{
		Index:       len(bk.Chain),
		Timestamp:   timestamp,
		Transaction: bk.Transactions,
		Proof:       proof,
		PrevHash:    prevHash,
	}
	newBlock.Hash = HashGenerate(newBlock)
	bk.Chain = append(bk.Chain, newBlock)
	bk.Transactions = []Transaction{}
}

func (bc *Blockchain) AddTransaction(sender, recipient string, amount float64) string {
	txn := Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	}
	txn.HashID = HashID(txn)
	bc.Transactions = append(bc.Transactions, txn)
	return txn.HashID
}

func (bk *Blockchain) createGenesisBlock() {
	genesisBlock := Block{
		Index:       0,
		Timestamp:   time.Now().Unix(),
		Transaction: []Transaction{},
		Proof:       10,
		PrevHash:    "0",
	}
	genesisBlock.Hash = HashGenerate(genesisBlock)
	bk.Chain = append(bk.Chain, genesisBlock)
}

func validNonce(lastblock Block, proof int, transaction []Transaction, candidateTime int64) bool {
	tempBlock := Block{
		Index:       lastblock.Index,
		Timestamp:   lastblock.Timestamp,
		Transaction: transaction,
		Proof:       proof,
		PrevHash:    lastblock.PrevHash,
	}
	hash := HashGenerate(tempBlock)
	fmt.Printf("%s", hash)
	return hash[:4] == "0000"

}

func (bk *Blockchain) proofOfWork() (int, int64) {
	lastBlock := bk.Chain[len(bk.Chain)-1]
	candiateTime := time.Now().Unix()
	proof := 0

	for !validNonce(lastBlock, proof, bk.Transactions, candiateTime) {
		proof++
	}
	return proof, candiateTime
}

func main() {
	bc := NewBlockchain()
	bc.AddTransaction("Alice", "Bob", 50)
	bc.AddTransaction("Bob", "Charlie", 25)

	start := time.Now()

	proof, candidateTimestamp := bc.proofOfWork()

	duration := time.Since(start)
	fmt.Printf("Proof of Work знайдено: %d (час виконання: %s)\n", proof, duration)

	previousHash := bc.Chain[len(bc.Chain)-1].Hash
	bc.AddBlock(proof, candidateTimestamp, previousHash)

	fmt.Println("Blockchain:", bc.Chain)
}
