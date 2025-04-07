package po

import (
	"gorm.io/gorm"
)

// Task 运行流水记录
type User struct {
	gorm.Model
	UserName string `gorm:"column:user_name;NOT NULL"` // 用户名
	Password string `gorm:"column:password;NOT NULL"`  // 密码
}

func (t *User) TableName() string {
	return "user"
}
