package repository

import (
	"errors"
	"fmt"
	apitypes "lab4-swag/internal/app/api_types"
	"lab4-swag/internal/app/ds"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (r *Repository) GetUserByID(id uuid.UUID) (ds.User, error) {
	user := ds.User{}
	sub := r.db.Where("id = ?", id).Find(&user)
	if sub.Error != nil {
		return ds.User{}, sub.Error
	}
	if sub.RowsAffected == 0 {
		return ds.User{}, ErrNotFound
	}
	err := sub.First(&user).Error
	if err != nil {
		return ds.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByLogin(login string) (ds.User, error) {
	user := ds.User{}
	sub := r.db.Where("login = ?", login).Find(&user)
	if sub.Error != nil {
		return ds.User{}, sub.Error
	}
	if sub.RowsAffected == 0 {
		return ds.User{}, ErrNotFound
	}
	err := sub.First(&user).Error
	if err != nil {
		return ds.User{}, err
	}
	return user, nil
}

func (r *Repository) CreateUser(userJSON apitypes.UserJSON) (ds.User, error) {
	user := apitypes.UserFromJSON(userJSON)
	if user.Login == "" {
		return ds.User{}, errors.New("login is empty")
	}
	if user.Password == "" {
		return ds.User{}, errors.New("password is empty")
	}
	if _, err := r.GetUserByLogin(user.Login); err == nil {
		return ds.User{}, errors.New("user already exists")
	}

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return ds.User{}, err
	}
	user.Password = hashedPassword

	user.ID = uuid.New()

	sub := r.db.Create(&user)
	if sub.Error != nil {
		return ds.User{}, sub.Error
	}
	return user, nil
}

func (r *Repository) ChangeProfile(login string, userJSON apitypes.UserJSON) (ds.User, error) {
	currUser, err := r.GetUserByLogin(login)
	if err != nil {
		return ds.User{}, err
	}

	if userJSON.Login != "" {
		currUser.Login = userJSON.Login
	}

	if userJSON.Password != "" {
		hashedPassword, err := HashPassword(userJSON.Password)
		if err != nil {
			return ds.User{}, err
		}
		currUser.Password = hashedPassword
	}

	if userJSON.IsModerator && !currUser.IsModerator {
		userJSON.IsModerator = false
	}
	currUser.IsModerator = userJSON.IsModerator

	err = r.db.Save(&currUser).Error
	if err != nil {
		return ds.User{}, err
	}
	return currUser, nil
}

func (r *Repository) SignIn(userJSON apitypes.UserJSON) (string, error) {
	user, err := r.GetUserByLogin(userJSON.Login)
	if err != nil {
		return "", err
	}

	if !CheckPasswordHash(userJSON.Password, user.Password) {
		return "", errors.New("invalid password")
	}

	token, err := GenerateToken(user.ID, user.IsModerator)
	if err != nil {
		return "", err
	}

	return token, nil
}

func GenerateToken(id uuid.UUID, isModerator bool) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["user_id"] = id.String()
	claims["is_moderator"] = isModerator
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	fmt.Println(password)

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
