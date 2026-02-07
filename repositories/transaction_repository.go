package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0)

	for _, item := range items {
		var productPrice, stock int
		var productName string

		err := tx.QueryRow("SELECT name, price, stock FROM products WHERE id=$1", item.ProductID).Scan(&productName, &productPrice, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		subtotal := productPrice * item.Quantity
		totalAmount += subtotal

		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, item.ProductID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		details = append(details, models.TransactionDetail{
			ProductID:	item.ProductID,
			ProductName: productName,
			Quantity:	item.Quantity,
			Subtotal:	subtotal,
		})
	}

	var transactionID int
	var id int
	var createdAt time.Time
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id, created_at", totalAmount).Scan(&transactionID, &createdAt)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range details {
		details[i].TransactionID = transactionID
		err = tx.QueryRow("INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4) RETURNING id", transactionID, details[i].ProductID, details[i].Quantity, details[i].Subtotal).Scan(&id)
		details[i].ID = id
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}

	return &models.Transaction{
		ID:		transactionID, 
		TotalAmount: totalAmount,
		CreatedAt: createdAt,
		Details:	details,
	}, nil
}

func (repo *TransactionRepository) GetDailyReport() (*models.DailyReport, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	fmt.Println("startofDay:", startOfDay)
	fmt.Println("endOfDay:", endOfDay)

	var totalRevenue, totalTransaksi int
	err := repo.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*) 
		FROM transactions 
		WHERE created_at >= $1 AND created_at < $2
	`, startOfDay, endOfDay).Scan(&totalRevenue, &totalTransaksi)
	if err != nil {
		return nil, err
	}

	report := &models.DailyReport{
		TotalRevenue:   totalRevenue,
		TotalTransaksi: totalTransaksi,
	}

	var nama string
	var qtyTerjual int
	err = repo.db.QueryRow(`
		SELECT p.name, SUM(td.quantity) as qty
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE t.created_at >= $1 AND t.created_at < $2
		GROUP BY p.id, p.name
		ORDER BY qty DESC
		LIMIT 1
	`, startOfDay, endOfDay).Scan(&nama, &qtyTerjual)
	if err == nil {
		report.ProdukTerlaris = &models.TopProduct{
			Nama:       nama,
			QtyTerjual: qtyTerjual,
		}
	}

	return report, nil
}
