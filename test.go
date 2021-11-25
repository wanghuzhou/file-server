package main

import (
	_ "github.com/lib/pq"
)

func sqlTest() {

	sep := "----------\n"
	println(sep, "*sqlOpen")

	sqlSelect()
	println(sep, "*sqlSelect")

	sqlInsert()
	sqlSelect()
	println(sep, "*sqlInsert")

	sqlUpdate()
	sqlSelect()
	println(sep, "*sqlUpdate")

	//sqlDelete()
	sqlSelect()
	println(sep, "*sqlDelete")

	sqlClose()
	println(sep, "*sqlClose")
}

type UserInfo struct {
	uid        int
	username   string
	department string
	created    string
}
