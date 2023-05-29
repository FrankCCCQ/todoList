package repository

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"user/internal/service"
)

type User struct {
	UserID         uint   `gorm:"primarykey"`
	UserName       string `gorm:"unique"`
	NickName       string
	PasswordDigest string //密文
}

const (
	PasswordCost = 12 // 密码加密难度
)

func (user *User) CheckUserExist(req *service.UserRequest) bool {
	if err := DB.Where("user_name=?", req.UserName).First(&user).Error; err == gorm.ErrRecordNotFound {
		return false
	}
	return true
}

// ShowUserInfo : 获取用户信息
func (user *User) ShowUserInfo(req *service.UserRequest) (err error) {
	if exist := user.CheckUserExist(req); exist {
		return nil
	}
	return errors.New("UserName Not Exist")
}

// CreateUser 创建用户
func (*User) CreateUser(req *service.UserRequest) (user User, err error) {
	// 检验用户是不是已经存在
	var count int64
	DB.Where("user_name=?", req.UserName).Count(&count)
	if count != 0 {
		return User{}, errors.New("UserName Exist")
	}
	user = User{
		UserName: req.UserName,
		NickName: req.NickName,
	}
	// 加密
	_ = user.SetPassword(req.Password)
	err = DB.Create(&user).Error
	return user, err
}

// SetPassword 加密
func (user *User) SetPassword(passwd string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(passwd), PasswordCost)
	if err != nil {
		return err
	}
	user.PasswordDigest = string(bytes)
	return nil
}

// CheckPassword 检验密码
func (user *User) CheckPassword(passwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(passwd))
	return err == nil
}

// BuildUser 序列化对象
func BuildUser(item User) *service.UserModel {
	userModel := service.UserModel{
		UserID:   uint32(item.UserID),
		UserName: item.UserName,
		NickName: item.NickName,
	}
	return &userModel
}
