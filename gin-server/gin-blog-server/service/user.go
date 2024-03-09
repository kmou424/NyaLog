package service

import (
	"NyaLog/gin-blog-server/middleware"
	"NyaLog/gin-blog-server/model"
	"NyaLog/gin-blog-server/utils"
	"NyaLog/gin-blog-server/utils/errmsg"
	"math/rand"
	"regexp"
	"time"
)

// 判断用户是否存在
func UserExist() (bool, int) {
	return model.UserExist()
}

// 创建用户
func CreateUser(user *model.User) ([]byte, int) {
	// 验证用户是否已经存在
	userexist, err := UserExist()
	if err == errmsg.ERROR {
		return nil, errmsg.ERROR
	}
	// 若存在则判断是否验证，若验证则返回用户存在，若未验证则删除该用户重新注册
	if userexist {
		uservalidate, err := model.UserValidate()
		if err == errmsg.ERROR {
			return nil, errmsg.ERROR
		}
		if uservalidate {
			return nil, errmsg.UserExist
		} else {
			model.DeleteUser()
		}
	}
	// 正则判断密码是否含有数字、大小写字母、标点符号；密码长度需要大于6
	re := regexp.MustCompile(`[0-9]+[a-zA-Z]+[!@#$%^&*().]+`)
	match := re.MatchString(user.Password)
	if user.Uid == "" || user.Username == "" || !match || len(user.Password) <= utils.PasswordMinLen || user.Email == "" {
		return nil, errmsg.UserInfoError
	}

	// 账户信息通过后，开始认证2FA
	secret, url, err := middleware.GenerateTwoFA(user.Uid)
	user.Secret = secret
	if err == errmsg.GenerateSecretFailed {
		return nil, errmsg.GenerateQRFailed
	}
	qrcode, err := middleware.GenerateQRcode(url, 256)
	if err == errmsg.GenerateQRFailed {
		return nil, errmsg.GenerateQRFailed
	}

	// 用户注册
	err = model.NewUser(user)
	if err == errmsg.ERROR {
		return nil, errmsg.ERROR
	}
	return qrcode, errmsg.SUCCESS
}

// 发送邮件验证码
func SendEmailCode() int {
	// 验证用户是否已经存在
	userexist, err := UserExist()
	if err == errmsg.ERROR {
		return err
	}
	if userexist {
		uservalidate, err := model.UserValidate()
		if err == errmsg.ERROR {
			return err
		}
		if uservalidate {
			return errmsg.UserExist
		}
	} else {
		return errmsg.UserNotExist
	}

	var user model.User
	user, err = model.SeleUser()
	if err == errmsg.ERROR {
		return err
	}
	// 生成邮件验证码并且发送邮件信息
	rand.NewSource(time.Now().UnixNano())
	code := make([]byte, 6)
	for i := 0; i < 6; i++ {
		code[i] = byte(rand.Intn(10) + '0')
	}
	// 编辑发送邮件的信息
	msg := []byte("From: " + "NyaLog" + "\r\n" +
		"To: " + user.Email + "\r\n" +
		"Subject: " + "NyaLog: Your verification code is:" + "\r\n" +
		"\r\n" +
		string(code))
	err = middleware.SendEmail(user.Email, msg)
	if err != errmsg.SUCCESS {
		return err
	}
	middleware.UserEmailCode(user.Uid, string(code))
	return errmsg.SUCCESS
}

// 注册用户时验证用户需要用到的结构体
type CheckUserToken struct {
	Emailcode string `json:"emailcode"`
	Twofacode string `json:"twofacode"`
	Userip    string `json:"userip"`
}

// 创建用户时的验证
func CheckUser(data *CheckUserToken, token string) int {
	// 验证用户是否已经存在
	userexist, err := UserExist()
	if err == errmsg.ERROR {
		return errmsg.ERROR
	}
	if userexist {
		uservalidate, err := model.UserValidate()
		if err == errmsg.ERROR {
			return errmsg.ERROR
		}
		if uservalidate {
			return errmsg.UserExist
		}
	} else {
		return errmsg.UserNotExist
	}

	var user model.User
	user, _ = model.SeleUser()

	// 验证用户输入的验证信息
	uid, _ := middleware.ValidateJWT(token, data.Userip)
	if middleware.GetCode(uid) != data.Emailcode {
		return errmsg.UserEmailCodeError
	}
	if middleware.Validate2FA(data.Twofacode, user.Secret) == errmsg.CodeError {
		return errmsg.CodeError
	}
	user.Validateuser = 1
	err = model.UpdateUser(user.Uid, &user)
	if err == errmsg.ERROR {
		return errmsg.ERROR
	}
	middleware.DeleteCode(uid)
	middleware.DeleteToken(uid)
	return errmsg.SUCCESS
}
