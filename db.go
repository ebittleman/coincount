package coincount

import (
	"context"
	"database/sql"
	"math/big"
	"time"
)

var SQLDbCreate = `
BEGIN TRANSACTION;

CREATE TABLE account (
	id integer PRIMARY KEY,
	name text
);

CREATE TABLE gl_transaction (
	id integer,
	account_id integer,
	debit integer,
	credit integer,
	memo text,
	timestamp integer,
	PRIMARY KEY (id, account_id),
	FOREIGN KEY (account_id) REFERENCES account (id)
);

CREATE TABLE item (
	id integer PRIMARY KEY AUTOINCREMENT,
	name text
);

CREATE TABLE inventory_transaction (
	id integer PRIMARY KEY AUTOINCREMENT,
	account_id integer,
	item_id integer,
	qty_in text,
	qty_out text,
	cost integer,
	memo text,
	timestamp integer,
	FOREIGN KEY (account_id) REFERENCES account (id),
	FOREIGN KEY (item_id) REFERENCES item (id)
);

CREATE TABLE vendor (
	id integer PRIMARY KEY AUTOINCREMENT,
	name text
);

CREATE TABLE purchase (
	id integer PRIMARY KEY AUTOINCREMENT,
	vendor_id integer,
	payable_acct_id integer,
	amount integer,
	timestamp integer,
	FOREIGN KEY (vendor_id) REFERENCES vendor (id),
	FOREIGN KEY (payable_acct_id) REFERENCES account (id)
);

CREATE TABLE purchase_item (
	purchase_id integer,
	item_id integer,
	inventory_account_id integer,
	qty text,
	cost integer,
	amount integer,
	PRIMARY KEY (purchase_id, item_id),
	FOREIGN KEY (purchase_id) REFERENCES purchase (id),
	FOREIGN KEY (item_id) REFERENCES item (id),
	FOREIGN KEY (inventory_account_id) REFERENCES account (id)
);

CREATE TABLE posted_purchase (
	purchase_id integer PRIMARY KEY,
	transaction_id integer,
	timestamp integer,
	FOREIGN KEY (purchase_id) REFERENCES purchase (id),
	FOREIGN KEY (transaction_id) REFERENCES gl_transaction (id)
);

COMMIT;
`

type AccountTable struct {
	DB *sql.DB
}

func (a AccountTable) Save(ctx context.Context, acct Account) error {
	_, err := a.DB.ExecContext(ctx,
		"INSERT INTO account(id, name) VALUES (?, ?)",
		acct.ID, acct.Name)
	return err
}

func (a AccountTable) Get(ctx context.Context, id int) (Account, error) {
	var (
		acct Account
		err  error
	)

	row := a.DB.QueryRowContext(ctx,
		"SELECT id, name FROM account WHERE id=?",
		id)

	err = row.Scan(&acct.ID, &acct.Name)

	return acct, err
}

type ItemTable struct {
	DB *sql.DB
}

func (i ItemTable) Save(ctx context.Context, item Item) error {
	_, err := i.DB.ExecContext(ctx,
		"INSERT INTO item(id, name) VALUES (?, ?)",
		item.ID, item.Name)
	return err
}

func (i ItemTable) Get(ctx context.Context, id int) (Item, error) {
	var (
		item Item
		err  error
	)

	row := i.DB.QueryRowContext(ctx,
		"SELECT id, name FROM item WHERE id=?",
		id)

	err = row.Scan(&item.ID, &item.Name)

	return item, err
}

type VendorTable struct {
	DB *sql.DB
}

func (v VendorTable) Save(ctx context.Context, vendor Vendor) error {
	_, err := v.DB.ExecContext(ctx,
		"INSERT INTO vendor(id, name) VALUES (?, ?)",
		vendor.ID, vendor.Name)
	return err
}

func (v VendorTable) Get(ctx context.Context, id int) (Vendor, error) {
	var (
		vendor Vendor
		err    error
	)

	row := v.DB.QueryRowContext(ctx,
		"SELECT id, name FROM vendor WHERE id=?",
		id)

	err = row.Scan(&vendor.ID, &vendor.Name)

	return vendor, err
}

type PurchaseTable struct {
	DB *sql.DB
	PurchaseItemTable
}

func (p PurchaseTable) Save(
	ctx context.Context,
	purchase Purchase,
) (int, error) {
	var purchaseID int
	tx, err := p.DB.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return purchaseID, err
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO purchase
		(vendor_id, payable_acct_id, amount, timestamp) VALUES 
		(?, ?, ?, ?)`,
		purchase.Vendor.ID,
		purchase.PayableAccount.ID,
		purchase.Amount,
		purchase.Date.UTC().Unix(),
	)

	if err != nil {
		return purchaseID, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return purchaseID, err
	}

	purchaseID = int(id)

	for _, item := range purchase.Items {
		if err := p.SaveItem(ctx, tx, purchaseID, item); err != nil {
			return purchaseID, err
		}
	}

	return purchaseID, tx.Commit()
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

func (p PurchaseTable) Get(ctx context.Context, id int) (Purchase, error) {
	row := p.DB.QueryRowContext(ctx, `
		SELECT 
		purchase.id, 
		purchase.vendor_id,
		vendor.name,
		purchase.payable_acct_id,
		account.id,
		purchase.amount,
		purchase.timestamp
		FROM purchase
		INNER JOIN vendor on vendor.id = purchase.vendor_id
		INNER JOIN account on account.id = purchase.payable_acct_id
		WHERE purchase.id=?`, id)

	return p.marshalFromScanner(ctx, row)
}

func (p PurchaseTable) marshalFromScanner(
	ctx context.Context,
	scanner Scanner,
) (Purchase, error) {
	var (
		purchase  Purchase
		timestamp int64
		err       error
	)

	if err = scanner.Scan(
		&purchase.ID,
		&purchase.Vendor.ID,
		&purchase.Vendor.Name,
		&purchase.PayableAccount.ID,
		&purchase.PayableAccount.Name,
		&purchase.Amount,
		&timestamp,
	); err != nil {
		return purchase, err
	}

	purchase.Date = time.Unix(timestamp, 0).UTC()
	purchase.Items, err = p.GetItems(ctx, p.DB, purchase.ID)

	return purchase, err
}

type PurchaseItemTable struct {
}

func (p PurchaseItemTable) SaveItem(
	ctx context.Context,
	tx *sql.Tx,
	purchaseID int,
	item PurchaseItem,
) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO purchase_item
			(purchase_id, item_id, inventory_account_id, qty, cost, amount) VALUES 
			(?, ?, ?, ?, ?, ?);`,
		purchaseID,
		item.Item.ID,
		item.InventoryAccount.ID,
		item.Qty.Text(16),
		item.Cost,
		item.Amount,
	)
	return err
}

func (p PurchaseItemTable) GetItems(ctx context.Context, db *sql.DB, purchaseID int) ([]PurchaseItem, error) {
	var items []PurchaseItem
	rows, err := db.QueryContext(ctx,
		`
		SELECT 
			purchase_item.item_id,
			item.name,
			purchase_item.inventory_account_id,
			account.name,
			purchase_item.qty,
			purchase_item.cost,
			purchase_item.amount
		FROM purchase_item
		INNER JOIN item on item.id = purchase_item.item_id
		INNER JOIN account on account.id = purchase_item.inventory_account_id
		WHERE purchase_item.purchase_id=?`, purchaseID)

	if err != nil {
		if rows != nil {
			rows.Close()
		}

		return nil, err
	}

	for rows.Next() && err == nil {
		qty := ""
		items = append(items, PurchaseItem{})
		i := len(items) - 1
		err = rows.Scan(
			&items[i].Item.ID,
			&items[i].Item.Name,
			&items[i].InventoryAccount.ID,
			&items[i].InventoryAccount.Name,
			&qty,
			&items[i].Cost,
			&items[i].Amount,
		)
		items[i].Qty = big.NewInt(0)
		items[i].Qty.SetString(qty, 16)
	}

	rows.Close()
	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return items, nil
}

type GLTransactionTable struct {
	DB *sql.DB
}

func (g GLTransactionTable) NextID(ctx context.Context) (int, error) {
	var nextNum int
	row := g.DB.QueryRowContext(
		ctx,
		"SELECT id FROM gl_transaction ORDER BY id DESC LIMIT 1;",
	)

	err := row.Scan(&nextNum)
	if err == sql.ErrNoRows {
		return 1, nil
	}

	if err != nil {
		return -1, err
	}

	return nextNum + 1, nil
}

func (g GLTransactionTable) Save(ctx context.Context, transactions []GLTransaction) error {
	tx, err := g.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	for _, transaction := range transactions {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO gl_transaction
				(id, account_id, debit, credit, memo, timestamp) VALUES 
				(?, ?, ?, ?, ?, ?)`,
			transaction.ID,
			transaction.Account.ID,
			transaction.Debit,
			transaction.Credit,
			transaction.Memo,
			transaction.Date.UTC().Unix(),
		); err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func (g GLTransactionTable) Get(ctx context.Context, id int) ([]GLTransaction, error) {
	var (
		transactions []GLTransaction
		timestamp    int64
		err          error
	)

	rows, err := g.DB.QueryContext(ctx, `
		SELECT 
			gl_transaction.id, 
			gl_transaction.account_id,
			account.name,
			gl_transaction.debit,
			gl_transaction.credit,
			gl_transaction.memo,
			gl_transaction.timestamp
		FROM gl_transaction
		INNER JOIN account ON account.id = gl_transaction.account_id
		WHERE id=?`, id)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	for rows.Next() {
		transactions = append(transactions, GLTransaction{})
		i := len(transactions) - 1
		err = rows.Scan(
			&transactions[i].ID,
			&transactions[i].Account.ID,
			&transactions[i].Account.Name,
			&transactions[i].Debit,
			&transactions[i].Credit,
			&transactions[i].Memo,
			&timestamp,
		)
		transactions[i].Date = time.Unix(timestamp, 0)
	}

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return transactions, err
}

type InventoryTransactionTable struct {
	DB *sql.DB
}

func (i InventoryTransactionTable) Save(ctx context.Context, transaction InventoryTransaction) (int, error) {
	res, err := i.DB.ExecContext(ctx, `
		INSERT INTO inventory_transaction
		(account_id, item_id, qty_in, qty_out, cost, memo, timestamp) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		// transaction.ID, <- autoincrement
		transaction.Account.ID,
		transaction.Item.ID,
		transaction.QtyIn.Text(16),
		transaction.QtyOut.Text(16),
		transaction.Cost,
		transaction.Memo,
		transaction.Date.UTC().Unix(),
	)
	if err != nil {
		return -1, nil
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return int(id), nil
}

func (i InventoryTransactionTable) Get(ctx context.Context, id int) (InventoryTransaction, error) {
	var (
		transaction InventoryTransaction
		timestamp   int64
		qtyIn       string
		qtyOut      string
		err         error
	)

	row := i.DB.QueryRowContext(ctx, `
		SELECT 
			inventory_transaction.id, 
			inventory_transaction.account_id,
			account.name,
			inventory_transaction.item_id,
			item.name,
			inventory_transaction.qty_in,
			inventory_transaction.qty_out,
			inventory_transaction.memo,
			inventory_transaction.timestamp
		FROM inventory_transaction 
		INNER JOIN account ON account.id=inventory_transaction.account_id
		INNER JOIN item ON item.id=inventory_transaction.item_id
		WHERE id=?`,
		id)

	err = row.Scan(
		&transaction.ID,
		&transaction.Account.ID,
		&transaction.Account.Name,
		&transaction.Item.ID,
		&transaction.Item.Name,
		&qtyIn,
		&qtyOut,
		&transaction.Memo,
		&timestamp,
	)

	transaction.QtyIn, _ = big.NewInt(0).SetString(qtyIn, 16)
	transaction.QtyOut, _ = big.NewInt(0).SetString(qtyIn, 16)

	transaction.Date = time.Unix(timestamp, 0).UTC()

	return transaction, err
}
