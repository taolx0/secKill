package service

import (
	"context"
	"errors"
	"secKill/oauth-service/model"
	"secKill/pb"
	"secKill/pkg/client"
)

var (
	//InvalidAuthentication = errors.New("invalid auth")
	InvalidUserInfo = errors.New("invalid user info")
)

// Service Define a service interface
type UserDetailsService interface {
	// Get UserDetails By username
	GetUserDetailByUsername(ctx context.Context, username, password string) (*model.UserDetails, error)
}

//UserService implement Service interface
type RemoteUserService struct {
	userClient client.UserClient
}

func (service *RemoteUserService) GetUserDetailByUsername(ctx context.Context, username, password string) (*model.UserDetails, error) {
	response, err := service.userClient.CheckUser(ctx, nil, &pb.UserRequest{
		Username: username,
		Password: password,
	})

	if err == nil {
		if response.UserId != 0 {
			return &model.UserDetails{
				UserId:   response.UserId,
				Username: username,
				Password: password,
			}, nil
		} else {
			return nil, InvalidUserInfo
		}
	}
	return nil, err
}

func NewRemoteUserDetailService() *RemoteUserService {
	userClient, _ := client.NewUserClient("user", nil, nil)
	return &RemoteUserService{
		userClient: userClient,
	}
}

// Middleware define service middleware
type Middleware func(Service) Service
