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
		_ = db.Close()
	}()
	key := "adasdsa"
	value := "asdbsadsd"
	err = db.Put(key, value, nil)
	if err != nil {
		panic(err)
	}

	newValue, err := db.Get(key)
	if err != nil {
		panic(err)
	}
	print(string(newValue))
}
