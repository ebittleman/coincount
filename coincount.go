package coincount

import (
	"errors"
	"fmt"
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
		Debit   float64
		Credit  float64
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
		QtyIn   float64
		QtyOut  float64
		Cost    float64
		Amount  float64
		Memo    string
	}

	Vendor struct {
		ID   int
		Name string
	}

	PurchaseItem struct {
		Item             Item
		InventoryAccount Account
		Qty              float64
		Cost             float64
		Amount           float64
	}

	Purchase struct {
		ID             int
		Date           time.Time
		Vendor         Vendor
		PayableAccount Account
		Amount         float64
		Items          []PurchaseItem
	}
)

func MiningPayout(date time.Time, qty float64, costOfElecricity float64) Purchase {
	return Purchase{
		Date:           date,
		Vendor:         ElectricCompany,
		PayableAccount: ElectricBill,
		Amount:         qty * costOfElecricity,
		Items: []PurchaseItem{
			{
				Item:             Ether,
				InventoryAccount: EthMain,
				Qty:              qty,
				Cost:             costOfElecricity,
				Amount:           qty * costOfElecricity,
			},
		},
	}
}

func PostPurchase(date time.Time, purchase Purchase, nextGLTransaction int) ([]InventoryTransaction, []GLTransaction) {
	var (
		inventoryTransactions []InventoryTransaction
		glTransactions        []GLTransaction
	)

	memo := fmt.Sprintf("PUR-%d", purchase.ID)
	for _, item := range purchase.Items {
		if item.Item.ID <= 0 {
			continue
		}

		qtyIn, qtyOut := item.Qty, 0.0
		if qtyIn < 0 {
			qtyOut = -1 * qtyIn
			qtyIn = 0
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

		debitAmount, creditAmount := item.Amount, 0.0
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

	debitAmount, creditAmount := 0.0, purchase.Amount
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

// TODO: continue here.
func PurchaseAssetWithEth(
	date time.Time,
	vendor Vendor,
	assetAccount Account,
	ethAccount Account,
	qty float64,
	fee float64,
	costOfEth float64,
) Purchase {
	return Purchase{
		Date:           date,
		Vendor:         vendor,
		PayableAccount: ethAccount,
		Amount:         costOfEth * (qty + fee),
		Items: []PurchaseItem{
			{
				InventoryAccount: EthTXFee,
				Qty:              fee,
				Cost:             costOfEth,
				Amount:           fee * costOfEth,
			},
			{
				InventoryAccount: assetAccount,
				Qty:              1,
				Cost:             costOfEth * qty,
				Amount:           costOfEth * qty,
			},
		},
	}
}

func CalcCost(transactions []InventoryTransaction, qty float64) (cost float64, err error) {
	if qty == 0 {
		return 0, nil
	}

	inQueue, outQueue := make(TransactionQueue, 0), make(TransactionQueue, 0)
	for _, transaction := range transactions {
		if transaction.QtyIn > 0 {
			inQueue.Enqueue(transaction)
		} else if transaction.QtyOut > 0 {
			outQueue.Enqueue(transaction)
		}
	}

	var inQty, outQty, price float64
	var currentIn, currentOut InventoryTransaction
	for {
		if inQty == 0 {
			currentIn, err = inQueue.Dequeue()
			if err != nil {
				return 0.0, errors.New("Out of Inventory")
			}
			inQty = currentIn.QtyIn
		}

		if outQty == 0 {
			price = 0
			currentOut, err = outQueue.Peek()
			if err != nil {
				outQty = qty
			} else {
				outQty = currentOut.QtyOut
			}
		}

		if outQty <= inQty {
			inQty -= outQty
			price += (outQty * currentIn.Cost)
			outQty = 0
			_, err = outQueue.Dequeue()
			if err != nil {
				return price / qty, nil
			}
		} else {
			outQty -= inQty
			price += (inQty * currentIn.Cost)
			inQty = 0
		}
	}
}

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
		err = errors.New("Empty Queue")
		return
	}

	transaction = tmp[len(tmp)-1]
	*t = tmp[:len(tmp)-1]
	return
}

func (t TransactionQueue) Peek() (transaction InventoryTransaction, err error) {
	if len(t) < 1 {
		err = errors.New("Empty Queue")
		return
	}
	transaction = t[len(t)-1]
	return
}
