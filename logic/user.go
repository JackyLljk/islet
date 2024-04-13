package logic

import (
	"islet/dao/mysql"
	"islet/models"
)

// SignUp Handle the registration process
func SignUp(fo models.RegisterForm) (err error) {
	err = mysql.Register(&models.User{
		Username: fo.Username,
		Password: fo.Password,
	})

	return
}

// Login Handle the login process
func Login(user *models.User) (err error) {
	err = mysql.Login(user)

	return
}
