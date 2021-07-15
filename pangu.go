package storage


import (
	`github.com/storezhang/pangu`
)

func init() {
	app := pangu.New()

	if err := app.Sets(
		NewDisk,
		New,
	); nil != err {
		panic(err)
	}
}