package payment

import "github.com/jmoiron/sqlx"

type Repository interface {
	// 定義 repository 方法
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}
