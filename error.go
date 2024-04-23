package SmartStashDB

import "errors"

var (
	ErrDatabaseIsUsing = errors.New("the database directory is used by another process")
)
