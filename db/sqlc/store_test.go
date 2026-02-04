package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	fmt.Println(">>before:", fromAccount.Balance, toAccount.Balance)

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := testStore.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)
	// check errors
	for range n {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		require.Equal(t, result.Transfer.FromAccountID, fromAccount.ID)
		require.Equal(t, result.Transfer.ToAccountID, toAccount.ID)
		require.Equal(t, result.Transfer.Amount, amount)
		require.NotZero(t, result.Transfer.ID)
		require.NotZero(t, result.Transfer.CreatedAt)

		_, err = testQueries.GetTransfer(context.Background(), result.Transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromAccount.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, toAccount.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// TODO: check balance

		// check accounts
		require.NotEmpty(t, result.FromAccount)
		require.Equal(t, result.FromAccount.ID, fromAccount.ID)

		require.NotEmpty(t, result.ToAccount)
		require.Equal(t, result.ToAccount.ID, toAccount.ID)

		// check account's balance
		fmt.Println(">>tx:", result.FromAccount.Balance, result.ToAccount.Balance)
		diff1 := fromAccount.Balance - result.FromAccount.Balance
		diff2 := result.ToAccount.Balance - toAccount.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		numOfTransfers := int(diff1 / amount)
		require.True(t, numOfTransfers >= 1 && numOfTransfers <= n)
		require.NotContains(t, existed, numOfTransfers)
		existed[numOfTransfers] = true
	}

	// check the final updated balances
	updatedFromAccount, err := testQueries.GetAccount(context.Background(), fromAccount.ID)
	require.NoError(t, err)

	updatedToAccount, err := testQueries.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)

	fmt.Println(">>after:", updatedFromAccount.Balance, updatedToAccount.Balance)
	require.Equal(t, fromAccount.Balance-amount*int64(n), updatedFromAccount.Balance)
	require.Equal(t, toAccount.Balance+amount*int64(n), updatedToAccount.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	fmt.Println(">>before:", fromAccount.Balance, toAccount.Balance)

	// run n concurrent transfer transactions
	n := 10
	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountId := fromAccount.ID
		toAccountId := toAccount.ID

		if i%2 == 1 {
			fromAccountId = toAccount.ID
			toAccountId = fromAccount.ID
		}

		go func() {
			_, err := testStore.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountId,
				ToAccountID:   toAccountId,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	// check errors
	for range n {
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balances
	updatedFromAccount, err := testQueries.GetAccount(context.Background(), fromAccount.ID)
	require.NoError(t, err)

	updatedToAccount, err := testQueries.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)

	fmt.Println(">>after:", updatedFromAccount.Balance, updatedToAccount.Balance)
	require.Equal(t, fromAccount.Balance, updatedFromAccount.Balance)
	require.Equal(t, toAccount.Balance, updatedToAccount.Balance)
}
