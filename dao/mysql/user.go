package mysql

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"islet/models"
	"islet/pkg/snowflake"
)

const secret = "El Psy Congroo"

func Register(user *models.User) (err error) {
	sqlStr := "select count(user_id) from user where username = ?"
	var count int64
	err = db.Get(&count, sqlStr, user.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}

	if count > 0 {
		return ErrorUserExit
	}

	userId, err := snowflake.GetId()
	if err != nil {
		return ErrorGetIDFailed
	}
	password := encryptPassword([]byte(user.Password))

	sqlStr = "insert into user(user_id, username, password) values (?,?,?)"
	_, err = db.Exec(sqlStr, userId, user.Username, password)
	return
}

func Login(user *models.User) (err error) {
	originalPassword := user.Password
	sqlStr := "select user_id, username, password from user where username = ?"
	err = db.Get(user, sqlStr, user.Username)

	// error
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}

	// user is not exist
	if errors.Is(err, sql.ErrNoRows) {
		return ErrorUserNotExit
	}

	// determine if the password is correct
	password := encryptPassword([]byte(originalPassword))
	if user.Password != password {
		return ErrorPasswordWrong
	}

	return
}

func encryptPassword(data []byte) (result string) {
	h := md5.New()
	h.Write([]byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func GetUserByID(idStr string) (user *models.User, err error) {
	user = new(models.User)
	sqlStr := `select user_id, username from  user where user_id = ?`
	err = db.Get(user, sqlStr, idStr)
	return
}
