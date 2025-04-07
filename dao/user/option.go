package task

import "gorm.io/gorm"

type Option func(*gorm.DB) *gorm.DB

func WithUserName(userName string) Option {
	return func(d *gorm.DB) *gorm.DB {
		return d.Where("user_name = ?", userName)
	}
}
