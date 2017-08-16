package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path"
	"time"

	"github.com/ebittleman/coincount"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ElectricityPerETH = float64(102.0)
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	homeDir, ok := os.LookupEnv("USERPROFILE")
	if !ok {
		log.Fatal("Could not locate home directory")
	}

	db, err := sql.Open("sqlite3", path.Join(homeDir, "db.sqlite"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// initDB(ctx, db)

	// insertPurchases(ctx, db)

	// readPurchase(ctx, db, 1)

	postPurchase(ctx, db, readPurchase(ctx, db, 2))
}

func readPurchase(ctx context.Context, db *sql.DB, id int) coincount.Purchase {
	table := coincount.PurchaseTable{
		DB: db,
	}
	purchase, err := table.Get(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(purchase)
	return purchase
}

func postPurchase(ctx context.Context, db *sql.DB, purchase coincount.Purchase) {

	inventoryTable := coincount.InventoryTransactionTable{
		DB: db,
	}

	glTable := coincount.GLTransactionTable{
		DB: db,
	}

	nextId, err := glTable.NextID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Next GL Transactions:", nextId)

	inv, gl := coincount.PostPurchase(purchase.Date, purchase, nextId)

	for _, transaction := range inv {
		id, err := inventoryTable.Save(ctx, transaction)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Recorded Inventory Transaction:", id)
	}

	log.Println("Inventory Transactions:", inv)
	log.Println("GL Transactions:", gl)
	if err = glTable.Save(ctx, gl); err != nil {
		log.Fatal(err)
	}
}

func insertPurchases(ctx context.Context, db *sql.DB) {
	type call struct {
		Date time.Time
		Qty  float64
		Cost float64
	}

	calls := make([]call, 0, 10)

	file, err := os.Open("data/mining.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&calls)
	if err != nil {
		log.Fatal(err)
	}

	for _, call := range calls {
		purchase := coincount.MiningPayout(call.Date, call.Qty, call.Cost)
		table := coincount.PurchaseTable{
			DB: db,
		}

		id, err := table.Save(ctx, purchase)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Registered Purchase:", id)

	}
}

func initDB(ctx context.Context, db *sql.DB) {
	_, err := db.Exec(coincount.SQLDbCreate)
	if err != nil {
		log.Fatal(err)
	}

	insertFixtures(ctx, db)
}

func insertFixtures(ctx context.Context, db *sql.DB) {
	var err error
	accountTable := coincount.AccountTable{
		DB: db,
	}
	for _, acct := range coincount.GLAccounts {
		err = accountTable.Save(ctx, acct)
		if err != nil {
			log.Println(err)
		}
	}

	inventoryTable := coincount.ItemTable{
		DB: db,
	}
	for _, item := range coincount.InventoryItems {
		err = inventoryTable.Save(ctx, item)
		if err != nil {
			log.Println(err)
		}
	}

	vendorTable := coincount.VendorTable{
		DB: db,
	}
	for _, vendor := range coincount.Vendors {
		err = vendorTable.Save(ctx, vendor)
		if err != nil {
			log.Println(err)
		}
	}
}
