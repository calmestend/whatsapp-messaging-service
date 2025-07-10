package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	badger "github.com/dgraph-io/badger/v4"
)

type Account struct {
	Token         string `json:"token"`
	WabID         string `json:"wabID"`
	PhoneNumberID string `json:"phoneNumber"`
}

func RegisterAccount(w http.ResponseWriter, r *http.Request) {
	opts := badger.DefaultOptions("./badger_data")
	db, err := badger.Open(opts)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var account Account
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&account); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if account.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}
	if account.PhoneNumber == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	err = db.Update(func(txn *badger.Txn) error {
		key := []byte("account:" + account.PhoneNumber)
		value := []byte(account.Token)

		return txn.Set(key, value)
	})

	if err != nil {
		log.Printf("Error saving to database: %v", err)
		http.Error(w, "Failed to save account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]string{
		"message": "Account registered successfully",
		"token":   account.Token,
		"phone":   account.PhoneNumber,
	}

	json.NewEncoder(w).Encode(response)
}
