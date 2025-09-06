package kvstore

// Example of using transactions with the KV store
//
// Transactions ensure that multiple operations happen atomically:
// - All operations succeed together, OR
// - All operations fail together (nothing is saved)
//
// This is useful when you need to:
// - Update multiple related values consistently
// - Transfer data between keys
// - Perform complex updates based on current values

func ExampleTransaction() {
	store := New[string, int]("mystore")
	
	// Simple transaction - all these happen together
	store.Transaction(func(tx *Tx[string, int]) error {
		tx.Set("apples", 5)
		tx.Set("oranges", 3)
		tx.Set("total", 8)
		return nil // Success - all saved
	})
	
	// Bank transfer example - atomic money transfer
	store.Transaction(func(tx *Tx[string, int]) error {
		// Read current balances
		aliceBalance, _ := tx.Get("alice_balance")
		bobBalance, _ := tx.Get("bob_balance")
		
		// Transfer 50 from Alice to Bob
		amount := 50
		
		// Update both balances
		tx.Set("alice_balance", aliceBalance - amount)
		tx.Set("bob_balance", bobBalance + amount)
		
		return nil // Both changes saved together
	})
	
	// If error is returned, nothing is saved
	store.Transaction(func(tx *Tx[string, int]) error {
		tx.Set("temp", 999) // This won't be saved
		return errors.New("cancel") // Rollback everything
	})
}