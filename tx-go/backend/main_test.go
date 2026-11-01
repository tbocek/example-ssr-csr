package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestTransferStockWithTransaction_Success(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	fromID, toID, stock := 1, 2, 3

	// Mock transaction
	mock.ExpectBegin()
	mock.ExpectExec("SELECT id FROM products").
		WithArgs(fromID, toID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Mock SELECT from source product
	mock.ExpectQuery("SELECT stock FROM products WHERE id = \\$1").
		WithArgs(fromID).
		WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(10))

	// Mock UPDATE source product
	mock.ExpectExec("UPDATE products SET stock = stock - \\$1 WHERE id = \\$2").
		WithArgs(stock, fromID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT target product
	mock.ExpectQuery("SELECT stock FROM products WHERE id = \\$1").
		WithArgs(toID).
		WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(5))

	// Mock UPDATE target product
	mock.ExpectExec("UPDATE products SET stock = stock \\+ \\$1 WHERE id = \\$2").
		WithArgs(stock, toID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// Act
	err = transferStockWithTransaction(fromID, toID, stock)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestTransferStockWithTransaction_InsufficientStock(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	fromID, toID, stock := 1, 2, 10

	mock.ExpectBegin()
	mock.ExpectExec("SELECT id FROM products").
		WithArgs(fromID, toID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Mock SELECT from source product with insufficient stock
	mock.ExpectQuery("SELECT stock FROM products WHERE id = \\$1").
		WithArgs(fromID).
		WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(5))

	mock.ExpectRollback()

	// Act
	err = transferStockWithTransaction(fromID, toID, stock)

	// Assert
	if err == nil {
		t.Error("expected error for insufficient stock, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestTransferStockWithTransaction_ExceedsMaximum(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	fromID, toID, stock := 1, 2, 10

	mock.ExpectBegin()
	mock.ExpectExec("SELECT id FROM products").
		WithArgs(fromID, toID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Mock SELECT from source product
	mock.ExpectQuery("SELECT stock FROM products WHERE id = \\$1").
		WithArgs(fromID).
		WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(20))

	// Mock UPDATE source product
	mock.ExpectExec("UPDATE products SET stock = stock - \\$1 WHERE id = \\$2").
		WithArgs(stock, fromID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT target product with 95 units (would exceed 100)
	mock.ExpectQuery("SELECT stock FROM products WHERE id = \\$1").
		WithArgs(toID).
		WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(95))

	mock.ExpectRollback()

	// Act
	err = transferStockWithTransaction(fromID, toID, stock)

	// Assert
	if err == nil {
		t.Error("expected error for exceeding maximum, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestTransferStockWithoutTransaction_FailsAfterDeduction(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	fromID, toID, stock := 1, 2, 3

	// Mock UPDATE source product (succeeds)
	mock.ExpectExec("UPDATE products SET stock = stock - \\$1 WHERE id = \\$2").
		WithArgs(stock, fromID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// The function always returns error (simulated failure)
	// No second UPDATE is expected

	// Act
	err = transferStockWithoutTransaction(fromID, toID, stock)

	// Assert
	if err == nil {
		t.Error("expected error from simulated failure, got nil")
	}

	// This demonstrates the problem: first UPDATE committed, but function failed
	// In real scenario, stock would be deducted but never added
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMoveStockWithTransaction_Success(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	productID := 2

	mock.ExpectBegin()

	// Mock SELECT target product
	mock.ExpectQuery("SELECT stock FROM products WHERE id = \\$1").
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(5))

	// Mock UPDATE target product (deduct)
	mock.ExpectExec("UPDATE products SET stock = stock - 1 WHERE id = \\$1").
		WithArgs(productID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock UPDATE product 1 (add)
	mock.ExpectExec("UPDATE products SET stock = stock \\+ 1 WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// Act
	err = moveStockWithTransaction(productID)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMoveStockWithTransaction_NoStockToTransfer(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	productID := 2

	mock.ExpectBegin()

	// Mock SELECT target product with 0 units
	mock.ExpectQuery("SELECT stock FROM products WHERE id = \\$1").
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(0))

	mock.ExpectRollback()

	// Act
	err = moveStockWithTransaction(productID)

	// Assert
	if err == nil {
		t.Error("expected error for no stock, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
