package usecase

import (
	"fmt"
	repo "go-empty/pkg/app/user/repo/mysql"
	"go-empty/pkg/models"
	"go-empty/pkg/token"

	"golang.org/x/crypto/bcrypt"
)

var userRepo repo.User

type User struct {
}

func NewUser() *User {
	return &User{}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (*User) Register(user *models.User) (*models.User, error) {

	pw, err := HashPassword(user.Password)
	fmt.Printf("len %s %d\n", pw, len(pw))
	if err != nil {
		return nil, err
	}
	user.Password = pw
	return userRepo.CreateUser(user)
}

func (*User) Login(name, password string) (*LoginUser, error) {

	user, err := userRepo.GetUser(name)
	if err != nil {
		return nil, fmt.Errorf("can't not find user")
	}

	hashedPassword := password

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("invalid password : %s", err.Error())
	}

	lu := &LoginUser{
		User:  *user,
		Token: token.GenUUIDv4String(),
	}
	return lu, nil
}

type LoginUser struct {
	models.User
	Token string
}
