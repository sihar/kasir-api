package services

import (
	"kasir-api/models"
	"kasir-api/repositories"
	"time"
)

type TransactionService struct {
	repo *repositories.TransactionRepository
}

func NewTransactionService(repo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Checkout(items []models.CheckoutItem, useLock bool) (*models.Transaction, error) {
	return s.repo.CreateTransaction(items)
}


func (s *TransactionService) GetDailyReport() (*models.DailyReport, error) {
	return s.repo.GetDailyReport()
}

func (s *TransactionService) GetReportByDateRange(startDate, endDate string) (*models.DailyReport, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}
	end = end.Add(24 * time.Hour)
	return s.repo.GetReportByDateRange(start, end)
}
