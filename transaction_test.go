package kvstore_test

import (
	"errors"
	"testing"

	"github.com/xtdlib/kvstore"
)

// TestTransactionBasic shows basic transaction usage
func TestTransactionBasic(t *testing.T) {
	store := kvstore.New[string, int]("test_tx_basic")
	defer store.Clear()
	
	// Example 1: Simple successful transaction
	err := store.Transaction(func(tx *kvstore.Tx[string, int]) error {
		// These operations happen together
		if err := tx.Set("apple", 5); err != nil {
			return err
		}
		if err := tx.Set("banana", 3); err != nil {
			return err
		}
		if err := tx.Set("orange", 7); err != nil {
			return err
		}
		return nil // All changes are saved
	})
	
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
	
	// Verify all values were saved
	if store.Get("apple") != 5 {
		t.Error("apple should be 5")
	}
	if store.Get("banana") != 3 {
		t.Error("banana should be 3")
	}
	if store.Get("orange") != 7 {
		t.Error("orange should be 7")
	}
}

// TestTransactionRollback shows that failed transactions don't save any changes
func TestTransactionRollback(t *testing.T) {
	store := kvstore.New[string, int]("test_tx_rollback")
	defer store.Clear()
	
	// Set initial value
	store.Set("counter", 100)
	
	// Try a transaction that will fail
	err := store.Transaction(func(tx *kvstore.Tx[string, int]) error {
		// Try to update the counter
		if err := tx.Set("counter", 200); err != nil {
			return err
		}
		
		// Something goes wrong - return an error
		// This will cancel ALL changes in this transaction
		return errors.New("something went wrong")
	})
	
	// Transaction should have failed
	if err == nil {
		t.Fatal("Expected transaction to fail")
	}
	
	// Counter should still be 100 (not 200) because transaction was rolled back
	if store.Get("counter") != 100 {
		t.Errorf("counter should still be 100, got %d", store.Get("counter"))
	}
}

// TestTransactionReadWrite shows reading and writing in same transaction
func TestTransactionReadWrite(t *testing.T) {
	store := kvstore.New[string, int]("test_tx_readwrite")
	defer store.Clear()
	
	// Set initial values
	store.Set("balance_alice", 100)
	store.Set("balance_bob", 50)
	
	// Transfer money from Alice to Bob in a transaction
	err := store.Transaction(func(tx *kvstore.Tx[string, int]) error {
		// Read Alice's balance
		aliceBalance, err := tx.Get("balance_alice")
		if err != nil {
			return err
		}
		
		// Read Bob's balance
		bobBalance, err := tx.Get("balance_bob")
		if err != nil {
			return err
		}
		
		transferAmount := 30
		
		// Check if Alice has enough money
		if aliceBalance < transferAmount {
			return errors.New("insufficient funds")
		}
		
		// Update both balances atomically
		if err := tx.Set("balance_alice", aliceBalance-transferAmount); err != nil {
			return err
		}
		if err := tx.Set("balance_bob", bobBalance+transferAmount); err != nil {
			return err
		}
		
		return nil // Commit the transfer
	})
	
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}
	
	// Verify the transfer worked
	if store.Get("balance_alice") != 70 {
		t.Errorf("Alice should have 70, got %d", store.Get("balance_alice"))
	}
	if store.Get("balance_bob") != 80 {
		t.Errorf("Bob should have 80, got %d", store.Get("balance_bob"))
	}
}

// TestTransactionIsolation shows that changes in transaction aren't visible outside until commit
func TestTransactionIsolation(t *testing.T) {
	store := kvstore.New[string, int]("test_tx_isolation")
	defer store.Clear()
	
	store.Set("value", 1)
	
	// Start a transaction that will fail
	store.Transaction(func(tx *kvstore.Tx[string, int]) error {
		// Change value inside transaction
		tx.Set("value", 999)
		
		// Value inside transaction should be 999
		val, _ := tx.Get("value")
		if val != 999 {
			t.Error("Inside transaction, value should be 999")
		}
		
		// But outside the transaction (using store directly), it's still 1
		// because transaction hasn't committed yet
		outsideVal := store.Get("value")
		if outsideVal != 1 {
			t.Error("Outside transaction, value should still be 1")
		}
		
		// Return error to rollback
		return errors.New("rollback")
	})
	
	// After rollback, value should still be 1
	if store.Get("value") != 1 {
		t.Error("After rollback, value should be 1")
	}
}

// TestTransactionForEach shows iterating in a transaction
func TestTransactionForEach(t *testing.T) {
	store := kvstore.New[string, int]("test_tx_foreach")
	defer store.Clear()
	
	// Set initial values
	store.Set("a", 1)
	store.Set("b", 2)
	store.Set("c", 3)
	
	err := store.Transaction(func(tx *kvstore.Tx[string, int]) error {
		// Collect all key-value pairs first
		updates := make(map[string]int)
		err := tx.ForEach(func(key string, value int) error {
			updates[key] = value * 2
			return nil
		})
		if err != nil {
			return err
		}
		
		// Then update them
		for key, value := range updates {
			if err := tx.Set(key, value); err != nil {
				return err
			}
		}
		return nil
	})
	
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
	
	// Check all values were doubled
	if store.Get("a") != 2 {
		t.Error("a should be 2")
	}
	if store.Get("b") != 4 {
		t.Error("b should be 4")
	}
	if store.Get("c") != 6 {
		t.Error("c should be 6")
	}
}