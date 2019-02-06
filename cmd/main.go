package main

import (
	service "github.com/egorka-gh/zbazar/zsync/cmd/service"
	_ "github.com/go-sql-driver/mysql"
)

//go run main.go -id=00 -mysql="root:3411@tcp(127.0.0.1:3306)/pshdata" -folder="D:\Buffer\zexch"

func main() {
	service.Run()
}
