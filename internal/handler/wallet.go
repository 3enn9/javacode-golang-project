package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"golang-wallet/internal/db"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type WalletOperationRequest struct {
	WalletID      string `json:"walletId"`
	OperationType string `json:"operationType"`
	Amount        float64 `json:"amount"`
}

func isValidUUID(u string)bool{
	_, err := uuid.Parse(u)
	return err == nil
}

var (
	ErrInvalidUUID         = errors.New("invalid wallet ID")
	ErrInvalidOperation    = errors.New("invalid operation type")
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrWalletNotFound      = sql.ErrNoRows
)

func WalletOperationHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request WalletOperationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if !isValidUUID(request.WalletID) {
			http.Error(w, ErrInvalidUUID.Error(), http.StatusBadRequest)
			return
		}

		if request.OperationType != "DEPOSIT" && request.OperationType != "WITHDRAW" {
			http.Error(w, ErrInvalidOperation.Error(), http.StatusBadRequest)
			return
		}

		if request.Amount <= 0 {
			http.Error(w, ErrInvalidAmount.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Processing operation: %+v", request)


		var err error
		switch request.OperationType {
		case "DEPOSIT":
			err = db.Deposit(database, request.WalletID, request.Amount)
		case "WITHDRAW":
			err = db.Withdraw(database, request.WalletID, request.Amount)
		}

		if err != nil {
			if errors.Is(err, db.ErrInsufficientFunds) {
				http.Error(w, "Insufficient funds", http.StatusBadRequest)
			} else if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Wallet not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to process operation", http.StatusInternalServerError)
				log.Println(err)
			}
			return
		}

		response := map[string]string{
			"message": "Operation successful",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func GetBalanceHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		walletID := r.URL.Path[len("/api/v1/wallets/"):]

		if !isValidUUID(walletID) {
			http.Error(w, ErrInvalidUUID.Error(), http.StatusBadRequest)
			return
		}
		
		log.Printf("Fetching balance for WalletID: %s", walletID)

		balance, err := db.GetBalance(database, walletID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Wallet not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to fetch wallet balance", http.StatusInternalServerError)
			}
			return
		}

		response := map[string]interface{}{
			"walletId": walletID,
			"balance":  balance,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}