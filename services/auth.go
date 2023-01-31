package service

import (
	"backend-api/cache"
	authdto "backend-api/dto/auth"
	"backend-api/models"
	"backend-api/pkg/bcrypt"
	jwtToken "backend-api/pkg/jwt"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
)

type AuthService interface {
	Register(request authdto.RegisterRequest) (*authdto.RegisterResponse, error)
	Login(request authdto.LoginRequest) (*authdto.LoginResponse, error)
	GetUserID(ID int) (*authdto.CheckAutResponse, error)
}

type authService struct {
	// AuthRepository repositories.AuthRepository
	AuthCache cache.AuthCache
	validator *validator.Validate
}

func NewAuthService(AuthCache cache.AuthCache) *authService {
	return &authService{AuthCache, validator.New()}
}

func (s *authService) Register(request authdto.RegisterRequest) (*authdto.RegisterResponse, error) {
	//check validation struct
	err := s.validator.Struct(request)
	if err != nil {
		return nil, err
	}

	password, err := bcrypt.HashingPassword(request.Password)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:     request.Name,
		Email:    request.Email,
		Password: password,
		Status:   "customer",
	}

	data, err := s.AuthCache.Register(user)
	if err != nil {
		return nil, err
	}

	response := authdto.RegisterResponse{
		Name:     data.Name,
		Email:    data.Email,
		Password: data.Password,
		Status:   data.Status,
	}

	return &response, nil
}

func (s *authService) Login(request authdto.LoginRequest) (*authdto.LoginResponse, error) {
	//check validation struct
	err := s.validator.Struct(request)
	if err != nil {
		return nil, err
	}

	user, err := s.AuthCache.Login(request.Email)
	fmt.Println(user)

	if err != nil {
		return nil, err
	}

	//check password for auth
	isValid := bcrypt.CheckPasswordHash(request.Password, user.Password)
	if !isValid {
		return nil, fmt.Errorf("wrong email or password")
	}

	//generate token login
	claims := jwt.MapClaims{}
	claims["id"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 2).Unix() // 2 hour expired

	token, errGenerateToken := jwtToken.GenerateToken(&claims)
	if errGenerateToken != nil {
		log.Println(errGenerateToken)
		fmt.Println("unauthorized")
		return nil, errGenerateToken

	}

	loginResponse := authdto.LoginResponse{
		Id:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Token:  token,
		Status: user.Status,
	}

	return &loginResponse, nil
}

func (s *authService) GetUserID(ID int) (*authdto.CheckAutResponse, error) {
	user, err := s.AuthCache.GetUserID(ID)
	if err != nil {
		return nil, err
	}

	response := authdto.CheckAutResponse{
		Status: user.Status,
	}

	return &response, nil
}