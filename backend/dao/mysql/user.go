package mysql

import (
	"backend/models"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
)

// 把每一步数据库操作封装成函数
// 待logic层根据业务需求调用

const secret = "wcc"

// encryptPassword 对密码进行加密
func encryptPassword(data []byte) (result string) {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum(data))
}

// CheckUserExist 检查指定用户名的用户是否存在
func CheckUserExist(username string) (error error) {
	sqlStr := `select count(user_id) from user where username = ?`
	var count int
	if err := db.Get(&count, sqlStr, username); err != nil {
		return err
	}
	if count > 0 {
		return errors.New(ErrorUserExit)
	}
	return
}

// InsertUser 注册业务-向数据库中插入一条新的用户
func InsertUser(user models.User) (error error) {
	// 对密码进行加密
	user.Password = encryptPassword([]byte(user.Password))
	// 执行SQL语句入库
	sqlstr := `insert into user(user_id,username,password,email,gender,avatar) values(?,?,?,?,?,?)`
	_, err := db.Exec(sqlstr, user.UserID, user.UserName, user.Password, user.Email, user.Gender, user.Avatar)
	return err
}

// UpdateUserAvatar 更新用户头像
func UpdateUserAvatar(userID string, avatarURL string) error {
	sqlStr := "UPDATE user SET avatar = ? WHERE user_id = ?"
	_, err := db.Exec(sqlStr, avatarURL, userID)
	return err
}

// Login 登录业务
func Login(user *models.User) (err error) {
	originPassword := user.Password // 记录一下原始密码(用户登录的密码)
	sqlStr := "select user_id, username, password,avatar,gender from user where username = ?"
	err = db.Get(user, sqlStr, user.UserName)
	// 查询数据库出错
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	// 用户不存在
	if err == sql.ErrNoRows {
		return errors.New(ErrorUserNotExit)
	}
	// 生成加密密码与查询到的密码比较
	password := encryptPassword([]byte(originPassword))
	if user.Password != password {
		return errors.New(ErrorPasswordWrong)
	}
	return nil
}
