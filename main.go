package main

import (
	_const "SmartStashDB/const"
	"SmartStashDB/storage"
)

func main() {
	options := storage.DefaultOptions

	options.DirPath = _const.ExecDir() + "/data"

	db, err := storage.OpenDB(options)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = db.close()
	}()

}
