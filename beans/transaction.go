package beans

import (
	"log"
	"time"
)

type Transaction struct {
	FromAccountNo string `json:"from_account_no"`
	ToAccountNo string `json:"to_account_no"`
	Amount float64 `json:"amount"`
	Timestamp int64 `json:"timestamp"`
}

func Temp(){
	xyz := Transaction{Timestamp: time.Now().UnixMilli()}
	log.Println(xyz)
}

func (t *Transaction) ConvertEpochToTimestamp(epoch int64) time.Time{
	return time.UnixMilli(epoch)
}