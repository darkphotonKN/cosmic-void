package commonhelpers

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

/**
* Database Utilities - Helper Functions
**/

/**
* Accepts a function that expects a transaction argument and helps wrap the function call
* with the initiation, passing, and error checking of that transaction.
**/
func ExecTx(db *sqlx.DB, fn func(tx *sqlx.Tx) error) (err error) {
	tx, txBeginErr := db.Beginx() // NOTE: use "beginx" for the **sqlx* version

	if txBeginErr != nil {
		fmt.Printf("Error when attempting to start transaction: %v\n", txBeginErr)
		return fmt.Errorf("Error when attempting to start transaction: %v", txBeginErr)
	}

	// handles panics and errors and rollback if they occur
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()

			fmt.Println("Transaction rolled back due to panic:", p)

			// re-throw panic after rollback
			panic(p)
		}

		if err != nil {
			fmt.Printf("Error during transaction, rolling back: Error:%v\n", err)
			tx.Rollback()
		}
	}()

	// call the function passed in and provide the transaction to it
	err = fn(tx)

	if err != nil {
		return err
	}

	// no error, safe to commit
	if commitErr := tx.Commit(); commitErr != nil {
		fmt.Printf("Failed to commit transaction, rolling back. Error: %s\n", commitErr)
		tx.Rollback() // rollback if commit fails

		return commitErr
	}

	return nil
}
