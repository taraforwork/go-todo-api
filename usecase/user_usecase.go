package usecase

import (
	"go-todo-api/models"
	"go-todo-api/repository"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type IUserUsecase interface {
	SignUp(user models.User) (models.UserResponse,error)
	Login(user models.User)(string, error) //JWTトークンを返す
}

type userUsecase struct {
	ur repository.IUserRepository
}

func NewUserUsecase(ur repository.IUserRepository) IUserUsecase{
	return &userUsecase{ur}
}

func (uu *userUsecase)SignUp(user models.User)(models.UserResponse,error){
	//パスワードをハッシュ化
	hash, err := 	bcrypt.GenerateFromPassword([]byte(user.Password),10) //第2引数は暗号の複雑さ
	if err != nil {
		return models.UserResponse{},err
	}
	newUser := models.User{Email: user.Email, Password: string(hash)}
	if err := uu.ur.CreateUser(&newUser); err != nil{
		return models.UserResponse{}, err
	}
	resUser := models.UserResponse{
		ID: newUser.ID,
		Email: newUser.Email,
	}
	return resUser,nil
}

func (uu *userUsecase) Login(user models.User)(string, error){
	//clientからくるemailがdbに存在するか確認する
	storedUser := models.User{}
	if err := uu.ur.GetUserByEmail(&storedUser,user.Email); err != nil{
		return "", err
	}
	//パスワードの一致確認
	err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		return "", err
	}
	//JWTトークンの生成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": storedUser.ID,
		"exp": time.Now().Add(time.Hour * 12).Unix(), //有効期限
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil{
		return "",err
	}
	return tokenString, nil
}