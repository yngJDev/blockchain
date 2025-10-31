package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"strings"
	"time"
)

type Block struct {
	Index       int           `json:"index"`
	Timestamp   int64         `json:"timestamp"`
	Transaction []Transaction `json:"transaction"`
	PrevHash    string        `json:"prevHash"`
	Proof       int           `json:"proof"`
	Hash        string        `json:"hash"`
}

type Transaction struct {
	Sender    string  `json:"sender"`
	HashID    string  `json:"hash_id"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
}

type Blockchain struct {
	Chain   []Block
	Mempool []Transaction
}

func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		Chain:   []Block{},
		Mempool: []Transaction{},
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
		Transaction: bk.Mempool,
		Proof:       proof,
		PrevHash:    prevHash,
	}
	newBlock.Hash = HashGenerate(newBlock)
	bk.Chain = append(bk.Chain, newBlock)
	bk.Mempool = []Transaction{}
	_ = bk.SaveMempool("mempool.json")
}

func (bc *Blockchain) AddTransaction(sender, recipient string, amount float64) string {
	txn := Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	}
	txn.HashID = HashID(txn)
	bc.Mempool = append(bc.Mempool, txn)

	return txn.HashID
}
func (bc *Blockchain) SaveMempool(filename string) error {
	data, err := json.MarshalIndent(bc.Mempool, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (bc *Blockchain) LoadMempool(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &bc.Mempool)
}

func (bc *Blockchain) addTransaction(c *gin.Context) {
	var txn Transaction
	if err := c.ShouldBind(&txn); err != nil {
		c.JSON(http.StatusBadRequest, createResponse("error", "Invalid data"))
		return
	}
	txID := bc.AddTransaction(txn.Sender, txn.Recipient, txn.Amount)
	responce := createResponse("success", map[string]string{"txID": txID})

	c.JSON(200, responce)
}

func (bk *Blockchain) createGenesisBlock() {
	genesisBlock := Block{
		Index:       0,
		Timestamp:   time.Now().Unix(),
		Transaction: []Transaction{},
		Proof:       10,
		PrevHash:    "prokofiev",
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
	return strings.HasPrefix(hash, "0000")

	//fmt.Printf("%s", hash)
	//
	//monthSuffix := "10"
	//return hash[:4] == "0000" &&
	//	strings.HasSuffix(hash, monthSuffix)

}

func (bk *Blockchain) proofOfWork() (int, int64) {
	lastBlock := bk.Chain[len(bk.Chain)-1]
	candiateTime := time.Now().Unix()
	proof := 0

	for !validNonce(lastBlock, proof, bk.Mempool, candiateTime) {
		proof++
	}
	return proof, candiateTime
}

func createResponse(status string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status": status,
		"data":   data,
	}
}

func (bc *Blockchain) mineBlockchain(c *gin.Context) {
	proof, candiateTime := bc.proofOfWork()
	prevHash := bc.Chain[len(bc.Chain)-1].Hash
	bc.AddBlock(proof, candiateTime, prevHash)
	responce := createResponse("success", map[string]interface{}{
		"index":     len(bc.Chain) - 1,
		"proof":     proof,
		"timestamp": candiateTime,
		"prevhash":  prevHash,
	})
	c.JSON(200, responce)
}

func (bc *Blockchain) getBlockchain(c *gin.Context) {
	responce := createResponse("success", bc.Chain)
	c.JSON(200, responce)
}

func main() {

	bc := NewBlockchain()
	_ = bc.LoadMempool("mempool.json")

	r := gin.Default()

	r.GET("/blockchain", bc.getBlockchain)
	r.POST("/transaction", bc.addTransaction)
	r.GET("/mine", bc.mineBlockchain)
	fmt.Println("Server start...")
	r.Run(":8080")
	/*r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run()
	*/
	/*bc := NewBlockchain()
	bc.AddTransaction("Alice", "Bob", 50)
	bc.AddTransaction("Bob", "Charlie", 25)

	start := time.Now()

	proof, candidateTimestamp := bc.proofOfWork()

	duration := time.Since(start)
	fmt.Printf("Proof of Work знайдено: %d (час виконання: %s)\n", proof, duration)

	previousHash := bc.Chain[len(bc.Chain)-1].Hash
	bc.AddBlock(proof, candidateTimestamp, previousHash)

	fmt.Println("Blockchain:", bc.Chain)*/

}
