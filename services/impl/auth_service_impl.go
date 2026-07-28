package impl

import (
	"context"
	"fmt"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/pkg/config"
	"error-logging/pkg/session"
	"error-logging/services"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

type authService struct {
	users    repository.UserRepository
	members  repository.OrgMemberRepository
	session  *session.Manager
	cfg      config.AuthConfig
	oauthCfg *oauth2.Config
}

func NewAuthService(
	users repository.UserRepository,
	members repository.OrgMemberRepository,
	sess *session.Manager,
	cfg config.AuthConfig,
) services.AuthService {
	var oauthCfg *oauth2.Config
	if cfg.GoogleEnabled() {
		oauthCfg = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Endpoint:     googleoauth.Endpoint,
			Scopes:       []string{"openid", "email", "profile"},
		}
	}
	return &authService{users: users, members: members, session: sess, cfg: cfg, oauthCfg: oauthCfg}
}

func (s *authService) CurrentUser(ctx context.Context, userID uint64) (*dbdto.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *authService) GoogleEnabled() bool { return s.oauthCfg != nil }

func (s *authService) GoogleAuthURL(state string) string {
	if s.oauthCfg == nil {
		return ""
	}
	return s.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline, oauth2.SetAuthURLParam("prompt", "select_account"))
}

func (s *authService) GoogleLogin(ctx context.Context, code string) (*dbdto.User, string, error) {
	if s.oauthCfg == nil {
		return nil, "", fmt.Errorf("google oauth not configured")
	}

	tok, err := s.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("exchange code: %w", err)
	}
	rawID, ok := tok.Extra("id_token").(string)
	if !ok || rawID == "" {
		return nil, "", fmt.Errorf("no id_token in google response")
	}

	payload, err := idtoken.Validate(ctx, rawID, s.cfg.GoogleClientID)
	if err != nil {
		return nil, "", fmt.Errorf("verify id_token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	if email == "" {
		return nil, "", fmt.Errorf("google account has no email")
	}
	if verified, _ := payload.Claims["email_verified"].(bool); !verified {
		return nil, "", fmt.Errorf("google email not verified")
	}
	name, _ := payload.Claims["name"].(string)

	return s.resolveAndIssue(ctx, email, name)
}

func (s *authService) resolveAndIssue(ctx context.Context, email, name string) (*dbdto.User, string, error) {
	user, err := s.users.FindOrCreateByEmail(ctx, email, name)
	if err != nil {
		return nil, "", fmt.Errorf("resolve user: %w", err)
	}

	if err := s.members.ActivateInvites(ctx, user.ID, user.Email); err != nil {
		return nil, "", fmt.Errorf("activate invites: %w", err)
	}

	token, err := s.session.Issue(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("issue token: %w", err)
	}
	return user, token, nil
}
