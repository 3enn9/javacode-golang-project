package db

import (
	"database/sql"
	"errors"
	"fmt"
	"golang-wallet/internal/model"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var ErrNotEnoughtMoney = errors.New("not enough money")

func InitDB() (*sql.DB, error) {

	if err := godotenv.Load("config.env"); err != nil {
		return nil, fmt.Errorf("error loading config.env file")
	}
	

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

    log.Println(dbHost, dbPort, dbUser, dbPassword)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to the database successfully.")
	return db, nil
}


func ApplyMigration(db *sql.DB, migrationFile string) error {
    sqlBytes, err := os.ReadFile(migrationFile)
    if err != nil {
        return fmt.Errorf("error reading migration file: %w", err)
    }

    sql := string(sqlBytes)
    _, err = db.Exec(sql)
    if err != nil {
        return fmt.Errorf("error executing migration: %w", err)
    }

    log.Println("Migration applied successfully.")
    return nil
}


func CreateWallet(db *sql.DB, wallet model.Wallet) error {
    query := `INSERT INTO wallets (id, balance) VALUES ($1, $2) RETURNING id, balance, created_at, updated_at`
    
    err := db.QueryRow(query, wallet.ID, wallet.Balance).Scan(&wallet.ID, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to insert wallet: %w", err)
    }
    
    log.Println("Wallet created successfully:", wallet)
    return nil
}

func GetBalance(database *sql.DB, walletID string) (float64, error) {
	query := `SELECT balance FROM wallets WHERE id = $1`
	var balance float64

    err := database.QueryRow(query, walletID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, sql.ErrNoRows
		}
		return 0, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	return balance, nil
}



func Deposit(database *sql.DB, walletID string, amount float64) error {
    tx, err := database.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()

    query := `UPDATE wallets SET balance = balance + $1 WHERE id = $2`
    result, err := tx.Exec(query, amount, walletID)
    if err != nil {
        return fmt.Errorf("failed to execute deposit query: %w", err)
    }

    rowsAffected, err := result.RowsAffected()

    if err != nil {
        return fmt.Errorf("failed to get affected rows: %w", err)
    }
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

func Withdraw(database *sql.DB, walletID string, amount float64) error {
    tx, err := database.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()

    var currentBalance float64
    checkBalanceQuery := `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`
    err = tx.QueryRow(checkBalanceQuery, walletID).Scan(&currentBalance)
    if err != nil {
        if err == sql.ErrNoRows {
            return sql.ErrNoRows 
        }
        return fmt.Errorf("failed to fetch balance: %w", err)
    }

    if currentBalance < amount {
        return ErrNotEnoughtMoney
    }

    updateQuery := `UPDATE wallets SET balance = balance - $1 WHERE id = $2`
    result, err := tx.Exec(updateQuery, amount, walletID)
    if err != nil {
        return fmt.Errorf("failed to execute withdraw query: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get affected rows: %w", err)
    }
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
