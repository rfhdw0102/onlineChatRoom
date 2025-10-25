package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

type user struct {
	Id       int    `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	//State    int    `db:"state"`
}

// ConnectDb 连接数据库
func ConnectDb() (err error) {
	dsn := "root:995812@tcp(localhost:3306)/onlinechatroom?charset=utf8mb4&parseTime=True&loc=Local"
	//连接数据库并尝试ping
	DB, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return fmt.Errorf("connect to mysql failed:%w", err)
	}
	return nil
}

// AddUserDb 注册用户
func AddUserDb(username string, password string) (err error) {
	sqlStr := "insert into user(username,password) values (?,?)"
	_, err = DB.Exec(sqlStr, username, password)
	if err != nil {
		return fmt.Errorf("AddUser failed:%w", err)
	}
	return nil
}

// SearchUserDb 查询用户
func SearchUserDb(username string) (pwd string, err error) {
	var u user
	sqlStr := "select id,username,password from user where username = ?"
	err = DB.Get(&u, sqlStr, username)
	if err != nil {
		return "", fmt.Errorf("SearchUser failed:%w", err)
	}
	return u.Password, nil
}

// Update 更新用户状态 0为离线，1为在线
//func Update(state int, username string) error {
//	sqlStr := "update user set state = ? where username = ?"
//	_, err := DB.Exec(sqlStr, state, username)
//	if err != nil {
//		return fmt.Errorf("update failed:%w", err)
//	}
//	return nil
//}
