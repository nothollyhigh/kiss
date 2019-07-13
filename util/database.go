package util

import (
	"database/sql"
	"fmt"
	"github.com/nothollyhigh/kiss/log"
)

// clear transaction
func ClearTx(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil && err != sql.ErrTxDone {
		fmt.Fprintf(log.DefaultLogger.Writer, log.LogWithFormater(log.LEVEL_ERROR, log.DefaultLogDepth+1, log.DefaultLogTimeLayout, "ClearTx failed: %v\n", nil))
	}
}
