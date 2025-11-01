package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestTransferStarsWithTransaction_Success(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	fromID, toID, stars := 1, 2, 3

	// Mock transaction
	mock.ExpectBegin()
	
	// Mock SELECT from source game
	mock.ExpectQuery("SELECT stars FROM games WHERE id = \\$1").
		WithArgs(fromID).
		WillReturnRows(sqlmock.NewRows([]string{"stars"}).AddRow(10))

	// Mock UPDATE source game
	mock.ExpectExec("UPDATE games SET stars = stars - \\$1 WHERE id = \\$2").
		WithArgs(stars, fromID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT target game
	mock.ExpectQuery("SELECT stars FROM games WHERE id = \\$1").
		WithArgs(toID).
		WillReturnRows(sqlmock.NewRows([]string{"stars"}).AddRow(5))

	// Mock UPDATE target game
	mock.ExpectExec("UPDATE games SET stars = stars \\+ \\$1 WHERE id = \\$2").
		WithArgs(stars, toID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// Act
	err = transferStarsWithTransaction(fromID, toID, stars)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestTransferStarsWithTransaction_InsufficientStars(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	fromID, toID, stars := 1, 2, 10

	mock.ExpectBegin()
	
	// Mock SELECT from source game with insufficient stars
	mock.ExpectQuery("SELECT stars FROM games WHERE id = \\$1").
		WithArgs(fromID).
		WillReturnRows(sqlmock.NewRows([]string{"stars"}).AddRow(5))

	mock.ExpectRollback()

	// Act
	err = transferStarsWithTransaction(fromID, toID, stars)

	// Assert
	if err == nil {
		t.Error("expected error for insufficient stars, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestTransferStarsWithTransaction_ExceedsMaximum(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	fromID, toID, stars := 1, 2, 10

	mock.ExpectBegin()
	
	// Mock SELECT from source game
	mock.ExpectQuery("SELECT stars FROM games WHERE id = \\$1").
		WithArgs(fromID).
		WillReturnRows(sqlmock.NewRows([]string{"stars"}).AddRow(20))

	// Mock UPDATE source game
	mock.ExpectExec("UPDATE games SET stars = stars - \\$1 WHERE id = \\$2").
		WithArgs(stars, fromID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock SELECT target game with 95 stars (would exceed 100)
	mock.ExpectQuery("SELECT stars FROM games WHERE id = \\$1").
		WithArgs(toID).
		WillReturnRows(sqlmock.NewRows([]string{"stars"}).AddRow(95))

	mock.ExpectRollback()

	// Act
	err = transferStarsWithTransaction(fromID, toID, stars)

	// Assert
	if err == nil {
		t.Error("expected error for exceeding maximum, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestTransferStarsWithoutTransaction_FailsAfterDeduction(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	fromID, toID, stars := 1, 2, 3

	// Mock UPDATE source game (succeeds)
	mock.ExpectExec("UPDATE games SET stars = stars - \\$1 WHERE id = \\$2").
		WithArgs(stars, fromID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// The function always returns error (simulated failure)
	// No second UPDATE is expected

	// Act
	err = transferStarsWithoutTransaction(fromID, toID, stars)

	// Assert
	if err == nil {
		t.Error("expected error from simulated failure, got nil")
	}

	// This demonstrates the problem: first UPDATE committed, but function failed
	// In real scenario, stars would be deducted but never added
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAddStarWithTransaction_Success(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	gameID := 2

	mock.ExpectBegin()
	
	// Mock SELECT target game
	mock.ExpectQuery("SELECT stars FROM games WHERE id = \\$1").
		WithArgs(gameID).
		WillReturnRows(sqlmock.NewRows([]string{"stars"}).AddRow(5))

	// Mock UPDATE target game (deduct)
	mock.ExpectExec("UPDATE games SET stars = stars - 1 WHERE id = \\$1").
		WithArgs(gameID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock UPDATE game 1 (add)
	mock.ExpectExec("UPDATE games SET stars = stars \\+ 1 WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// Act
	err = addStarWithTransaction(gameID)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAddStarWithTransaction_NoStarsToTransfer(t *testing.T) {
	// Arrange
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	db = mockDB

	gameID := 2

	mock.ExpectBegin()
	
	// Mock SELECT target game with 0 stars
	mock.ExpectQuery("SELECT stars FROM games WHERE id = \\$1").
		WithArgs(gameID).
		WillReturnRows(sqlmock.NewRows([]string{"stars"}).AddRow(0))

	mock.ExpectRollback()

	// Act
	err = addStarWithTransaction(gameID)

	// Assert
	if err == nil {
		t.Error("expected error for no stars, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}