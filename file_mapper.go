package main

import (
	"fmt"
)

func insertFile(file FileEntity) {

	if !conf.DB.Use {
		return
	}

	stmt, err := db.Prepare("INSERT INTO file(hash,path) VALUES($1,$2) RETURNING id")
	checkErr(err)

	res, err := stmt.Exec(file.Hash, file.Path)
	checkErr(err)
	affect, err := res.RowsAffected()
	checkErr(err)

	fmt.Println("rows affect:", affect)
}

func selectFile(file FileEntity) {

}

func sqlInsert() {
	//插入数据
	stmt, err := db.Prepare("INSERT INTO userinfo(username,departname,created) VALUES($1,$2,$3) RETURNING uid")
	checkErr(err)

	res, err := stmt.Exec("ficow", "软件开发部门", "2017-03-09")
	//这里的三个参数就是对应上面的$1,$2,$3了

	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	fmt.Println("rows affect:", affect)
}
func sqlDelete() {
	//删除数据
	stmt, err := db.Prepare("delete from userinfo where uid=$1")
	checkErr(err)

	res, err := stmt.Exec(1)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	fmt.Println("rows affect:", affect)
}
func sqlSelect() {
	//查询数据
	rows, err := db.Query("SELECT * FROM userinfo")
	checkErr(err)

	println("-----------")
	for rows.Next() {
		/*var uid int
		var username string
		var department string
		var created string
		err = rows.Scan(&uid, &username, &department, &created)*/
		userInfo := UserInfo{}
		err = rows.Scan(&userInfo.uid, &userInfo.username, &userInfo.department, &userInfo.created)
		checkErr(err)
		//fmt.Println("uid = ", uid, "\nname = ", username, "\ndep = ", department, "\ncreated = ", created, "\n-----------")
		fmt.Println("userInfo = ", userInfo)
	}
}
func sqlUpdate() {
	//更新数据
	stmt, err := db.Prepare("update userinfo set username=$1 where uid=$2")
	checkErr(err)

	res, err := stmt.Exec("ficow", 1)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	fmt.Println("rows affect:", affect)
}
func sqlClose() {
	db.Close()
}
