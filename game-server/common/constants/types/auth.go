package types

type Role int32

const (
	Role_ROLE_UNSPECIFIED Role = 0 // proto3 要求,代表未設定的角色
	Role_ROLE_PLAYER      Role = 1 // 對應資料庫的 'player'
	Role_ROLE_ADMIN       Role = 2 // 對應資料庫的 'admin'
)
