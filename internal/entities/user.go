package entities

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateAndLoginUserDTO struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

type UserDTO struct {
	ID       int    `form:"id" json:"id"`
	Username string `form:"username" json:"username"`
}

func (u *CreateAndLoginUserDTO) ToUser() User {
	return User{
		Username: u.Username,
		Password: u.Password,
	}
}

func (u *User) ToUserDTO() UserDTO {
	return UserDTO{
		ID:       int(u.ID),
		Username: u.Username,
	}
}
