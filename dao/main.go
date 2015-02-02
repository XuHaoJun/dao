package main

import (
    "runtime"
    "github.com/XuHaoJun/dao"
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    server, err := dao.NewServer()
    if err != nil {
        panic(err)
    }
    server.Run()
}
