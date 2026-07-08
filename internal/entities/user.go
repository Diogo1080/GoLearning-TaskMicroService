package entities

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserDTO struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

func (u *UserDTO) ToUser() User {
	return User{
		Username: u.Username,
		Password: u.Password,
	}
}
