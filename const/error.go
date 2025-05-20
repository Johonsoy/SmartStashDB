package _const

import (
	"errors"
)

var (
	ErrDatabaseIsUsing       = errors.New("the database directory is used by another process")
	ErrorDBClosed            = errors.New("the database is closed")
	ErrorReadOnlyBatch       = errors.New("the read-only batch exists")
	ErrorBatchCommited       = errors.New("the batch commited")
	ErrorKeyNotFound         = errors.New("key not found")
	ErrorKeyIsEmpty          = errors.New("the key is empty")
	ErrorFileExtError        = errors.New("segmentFileExt must not start with '.'")
	ErrorDataToLarge         = errors.New("data is too large")
	ErrorPendingSizeTooLarge = errors.New("pending size is too large")
)
