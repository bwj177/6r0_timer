package task

import (
	"context"
	"github.com/xiaoxuxiansheng/xtimer/common/model/po"
	"github.com/xiaoxuxiansheng/xtimer/pkg/mysql"
)

type UserDAO struct {
	client *mysql.Client
}

func NewUserDAO(client *mysql.Client) *UserDAO {
	return &UserDAO{
		client: client,
	}
}

func (t *UserDAO) CreateUser(ctx context.Context, user *po.User) error {
	return t.client.DB.WithContext(ctx).Create(user).Error
}

func (t *UserDAO) GetUser(ctx context.Context, opts ...Option) ([]*po.User, error) {
	db := t.client.DB.WithContext(ctx)
	for _, opt := range opts {
		db = opt(db)
	}

	var users []*po.User
	return users, db.Model(&po.User{}).Scan(&users).Error
}
