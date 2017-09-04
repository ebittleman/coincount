package coincount

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

type (
	Account struct {
		ID   int
		Name string
	}

	GLTransaction struct {
		ID      int
		Date    time.Time
		Account Account
		Debit   int64
		Credit  int64
		Memo    string
	}

	Item struct {
		ID   int
		Name string
	}

	InventoryTransaction struct {
		ID      int
		Date    time.Time
		Account Account
		Item    Item
		QtyIn   *big.Int
		QtyOut  *big.Int
		Cost    int64
		Amount  int64
		Memo    string
	}

	Vendor struct {
		ID   int
		Name string
	}

	PurchaseItem struct {
		Item             Item
		InventoryAccount Account
		Qty              *big.Int
		Cost             int64
		Amount           int64
	}

	Purchase struct {
		ID             int
		Date           time.Time
		Vendor         Vendor
		PayableAccount Account
		Amount         int64
		Items          []PurchaseItem
	}
)

func MiningPayout(date time.Time, qty *big.Int, costOfElecricity int64) Purchase {
	amt := multiplyRoundUp(qty, costOfElecricity)

	return Purchase{
		Date:           date,
		Vendor:         ElectricCompany,
		PayableAccount: ElectricBill,
		Amount:         amt,
		Items: []PurchaseItem{
			{
				Item:             Ether,
				InventoryAccount: EthMain,
				Qty:              qty,
				Cost:             costOfElecricity,
				Amount:           amt,
			},
		},
	}
}

func PostPurchase(date time.Time, purchase Purchase, nextGLTransaction int) ([]InventoryTransaction, []GLTransaction) {
	var (
		inventoryTransactions []InventoryTransaction
		glTransactions        []GLTransaction
		zero                  big.Int
	)

	memo := fmt.Sprintf("PUR-%d", purchase.ID)
	for _, item := range purchase.Items {
		if item.Item.ID <= 0 {
			continue
		}

		qtyIn, qtyOut := new(big.Int).Set(item.Qty), new(big.Int).Set(&zero)

		if qtyIn.Cmp(&zero) < 0 {
			qtyOut.Neg(qtyIn)
			qtyIn.Set(&zero)
		}

		inventoryTransactions = append(inventoryTransactions, InventoryTransaction{
			Date:    date,
			Account: item.InventoryAccount,
			Item:    item.Item,
			QtyIn:   qtyIn,
			QtyOut:  qtyOut,
			Cost:    item.Cost,
			Memo:    memo,
		})

		debitAmount, creditAmount := item.Amount, int64(0)
		if debitAmount < 0 {
			creditAmount = -1 * debitAmount
			debitAmount = 0
		}

		glTransactions = append(glTransactions, GLTransaction{
			ID:      nextGLTransaction,
			Date:    date,
			Account: item.InventoryAccount,
			Debit:   debitAmount,
			Credit:  creditAmount,
			Memo:    memo,
		})
	}

	debitAmount, creditAmount := int64(0), purchase.Amount
	if creditAmount < 0 {
		debitAmount = -1 * creditAmount
		creditAmount = 0
	}

	glTransactions = append(glTransactions, GLTransaction{
		ID:      nextGLTransaction,
		Date:    date,
		Account: purchase.PayableAccount,
		Debit:   debitAmount,
		Credit:  creditAmount,
		Memo:    memo,
	})

	return inventoryTransactions, glTransactions
}

// // TODO: continue here.
// func PurchaseAssetWithEth(
// 	date time.Time,
// 	vendor Vendor,
// 	assetAccount Account,
// 	ethAccount Account,
// 	qty *big.Int,
// 	fee *big.Int,
// 	costOfEth int64,
// ) Purchase {
// 	return Purchase{
// 		Date:           date,
// 		Vendor:         vendor,
// 		PayableAccount: ethAccount,
// 		Amount:         costOfEth * (qty + fee),
// 		Items: []PurchaseItem{
// 			{
// 				InventoryAccount: EthTXFee,
// 				Qty:              fee,
// 				Cost:             costOfEth,
// 				Amount:           fee * costOfEth,
// 			},
// 			{
// 				InventoryAccount: assetAccount,
// 				Qty:              1,
// 				Cost:             costOfEth * qty,
// 				Amount:           costOfEth * qty,
// 			},
// 		},
// 	}
// }

func CalcCost(transactions []InventoryTransaction, qty *big.Int) (cost int64, err error) {
	var zero big.Int
	if qty.Cmp(&zero) == 0 {
		return 0, nil
	}

	inQueue, outQueue := make(TransactionQueue, 0), make(TransactionQueue, 0)
	for _, transaction := range transactions {
		if transaction.QtyIn.Cmp(&zero) > 0 {
			inQueue.Enqueue(transaction)
		} else if transaction.QtyOut.Cmp(&zero) > 0 {
			outQueue.Enqueue(transaction)
		}
	}

	inQty, outQty := new(big.Int).Set(&zero), new(big.Int).Set(&zero)
	var price int64
	var currentIn, currentOut InventoryTransaction
	for {
		if inQty.Cmp(&zero) == 0 {
			currentIn, err = inQueue.Dequeue()
			if err != nil {
				return 0, errors.New("Out of Inventory")
			}
			inQty.Set(currentIn.QtyIn)
		}

		if outQty.Cmp(&zero) == 0 {
			price = 0
			currentOut, err = outQueue.Peek()
			if err != nil {
				outQty.Set(qty)
			} else {
				outQty.Set(currentOut.QtyOut)
			}
		}

		if outQty.Cmp(inQty) <= 0 {
			inQty.Sub(inQty, outQty)
			price += multiplyRoundUp(outQty, currentIn.Cost)
			outQty.Set(&zero)
			_, err = outQueue.Dequeue()
			if err != nil {
				return divideRound(price, qty), nil
			}
		} else {
			outQty.Sub(outQty, inQty)
			price += multiplyRoundUp(inQty, currentIn.Cost)
			inQty.Set(&zero)
		}
	}
}

var ErrEmptyQueue = errors.New("Empty Queue")

type TransactionQueue []InventoryTransaction

func (t *TransactionQueue) Enqueue(transaction InventoryTransaction) {
	tmp := *t
	tmp = append(tmp, transaction)

	for i := len(tmp) - 2; i >= 0; i-- {
		tmp[i+1] = tmp[i]
	}

	tmp[0] = transaction
	*t = tmp
}
func (t *TransactionQueue) Dequeue() (transaction InventoryTransaction, err error) {
	tmp := *t
	if len(tmp) < 1 {
		err = ErrEmptyQueue
		return
	}

	transaction = tmp[len(tmp)-1]
	*t = tmp[:len(tmp)-1]
	return
}

func (t TransactionQueue) Peek() (transaction InventoryTransaction, err error) {
	if len(t) < 1 {
		err = ErrEmptyQueue
		return
	}
	transaction = t[len(t)-1]
	return
}
