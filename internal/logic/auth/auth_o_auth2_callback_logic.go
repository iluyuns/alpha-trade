package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/iluyuns/alpha-trade/internal/model"
	"github.com/iluyuns/alpha-trade/internal/pkg/ctxval"
	"github.com/iluyuns/alpha-trade/internal/pkg/jwt"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthOAuth2CallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthOAuth2CallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) AuthOAuth2CallbackLogic {
	return AuthOAuth2CallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthOAuth2CallbackLogic) AuthOAuth2Callback(req *types.OAuth2CallbackResponse) (resp *types.LoginResponse, err error) {
	// 1. 解析并校验 State (还原 Provider 和 UID)
	provider, uid, err := l.parseState(req.State)
	if err != nil {
		return nil, fmt.Errorf("invalid oauth state: %w", err)
	}

	var conf *oauth2.Config
	switch provider {
	case "google":
		conf = &oauth2.Config{
			ClientID:     l.svcCtx.Config.OAuth.Google.ClientID,
			ClientSecret: l.svcCtx.Config.OAuth.Google.ClientSecret,
			RedirectURL:  l.svcCtx.Config.OAuth.Google.RedirectURL,
			Endpoint:     google.Endpoint,
		}
	case "github":
		conf = &oauth2.Config{
			ClientID:     l.svcCtx.Config.OAuth.Github.ClientID,
			ClientSecret: l.svcCtx.Config.OAuth.Github.ClientSecret,
			RedirectURL:  l.svcCtx.Config.OAuth.Github.RedirectURL,
			Endpoint:     github.Endpoint,
		}
	}

	// 2. 换取 Access Token
	token, err := conf.Exchange(l.ctx, req.Code)
	if err != nil {
		l.Errorf("OAuth2 exchange failed: %v", err)
		return nil, fmt.Errorf("auth: oauth exchange failed")
	}

	// 3. 拉取用户信息
	userInfo, err := l.fetchUserInfo(provider, token)
	if err != nil {
		l.Errorf("OAuth2 fetch user info failed: %v", err)
		return nil, fmt.Errorf("auth: failed to fetch profile")
	}

	// 4. 账户关联逻辑
	var user *model.Users
	if uid > 0 {
		// 情况 A: 绑定流程 (已登录用户关联第三方账号)
		user, err = l.svcCtx.UsersModel.FindOne(l.ctx, uid)
		if err != nil {
			return nil, fmt.Errorf("user not found for binding: %w", err)
		}

		// 执行绑定逻辑: 将第三方 ID 存入用户表
		switch provider {
		case "github":
			user.GithubId.String = userInfo.ID
			user.GithubId.Valid = true
		case "google":
			user.GoogleId.String = userInfo.ID
			user.GoogleId.Valid = true
		}
		if err := l.svcCtx.UsersModel.Update(l.ctx, user); err != nil {
			return nil, fmt.Errorf("failed to bind oauth account: %w", err)
		}
	} else {
		// 情况 B: 登录流程 (未登录用户通过已关联的第三方账号登录)
		user, err = l.svcCtx.UsersModel.FindOneByOAuth(l.ctx, provider, userInfo.ID)
		if err != nil {
			if err == sqlx.ErrNotFound {
				l.recordAccessLog(0, "OAUTH_LOGIN", "FAIL", "USER_NOT_FOUND")
				return nil, fmt.Errorf("account not found: %s account is not linked to any user", provider)
			}
			return nil, err
		}
	}

	// 5. 签发系统令牌 (ScopeBaseAuth)
	jwtToken, err := jwt.GenerateTokenWithIp(
		l.svcCtx.Config.Auth.AuthSecret,
		user.Id,
		jwt.ScopeBaseAuth,
		l.svcCtx.Config.Auth.BaseExpire,
		ctxval.GetIP(l.ctx),
	)
	if err != nil {
		return nil, err
	}

	l.recordAccessLog(user.Id, "OAUTH_LOGIN", "SUCCESS", provider)

	return &types.LoginResponse{
		Status: "success",
		Token:  jwtToken,
		User: types.UserInfo{
			Id:          user.Id,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
		},
	}, nil
}

func (l *AuthOAuth2CallbackLogic) parseState(stateStr string) (string, int64, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(stateStr)
	if err != nil {
		return "", 0, err
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 4 {
		return "", 0, fmt.Errorf("invalid state format")
	}

	provider := parts[0]
	uid, _ := strconv.ParseInt(parts[1], 10, 64)
	tsStr := parts[2]
	sig := parts[3]

	// 校验签名
	raw := fmt.Sprintf("%s:%d:%s", provider, uid, tsStr)
	h := hmac.New(sha256.New, []byte(l.svcCtx.Config.Auth.AuthSecret))
	h.Write([]byte(raw))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if sig != expectedSig {
		return "", 0, fmt.Errorf("state signature mismatch")
	}

	return provider, uid, nil
}

type providerProfile struct {
	ID     string
	Email  string
	Name   string
	Avatar string
}

func (l *AuthOAuth2CallbackLogic) fetchUserInfo(provider string, token *oauth2.Token) (*providerProfile, error) {
	client := oauth2.NewClient(l.ctx, oauth2.StaticTokenSource(token))
	var url string
	if provider == "google" {
		url = "https://www.googleapis.com/oauth2/v2/userinfo"
	} else {
		url = "https://api.github.com/user"
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	profile := &providerProfile{}
	if provider == "google" {
		profile.ID, _ = data["id"].(string)
		profile.Email, _ = data["email"].(string)
		profile.Name, _ = data["name"].(string)
		profile.Avatar, _ = data["picture"].(string)
	} else {
		// GitHub ID 可能是一个数字，需要转为 string
		if id, ok := data["id"].(float64); ok {
			profile.ID = strconv.FormatInt(int64(id), 10)
		} else if id, ok := data["id"].(string); ok {
			profile.ID = id
		}
		profile.Email, _ = data["email"].(string)
		if profile.Email == "" {
			// GitHub ID 作为兜底防止冲突
			profile.Email = fmt.Sprintf("%v@github.com", data["id"])
		}
		profile.Name, _ = data["name"].(string)
		if profile.Name == "" {
			profile.Name, _ = data["login"].(string)
		}
		profile.Avatar, _ = data["avatar_url"].(string)
	}
	return profile, nil
}

func (l *AuthOAuth2CallbackLogic) recordAccessLog(uid int64, action, status, reason string) {
	_, _ = l.svcCtx.UserAccessLogsModel.Insert(l.ctx, &model.UserAccessLogs{
		UserId:    uid,
		IpAddress: ctxval.GetIP(l.ctx),
		UserAgent: ctxval.GetUA(l.ctx),
		Action:    action,
		Status:    status,
		Reason:    reason,
	})
}
