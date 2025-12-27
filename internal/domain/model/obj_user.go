package model

import (
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// Role 定义用户角色
type Role string

const (
	RoleAdmin    Role = "ADMIN"
	RoleOperator Role = "OPERATOR"
	RoleViewer   Role = "VIEWER"
)

// User 用户实体，实现了 webauthn.User 接口
type User struct {
	ID          int64     `json:"id"`           // 数据库自增 ID
	UUID        string    `json:"uuid"`         // 对外公开的 UUID (WebAuthnID)
	Username    string    `json:"username"`     // 登录名
	DisplayName string    `json:"display_name"` // 显示名
	Role        Role      `json:"role"`         // 权限角色
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 聚合根关联：用户的 Passkeys 列表
	Credentials []WebAuthnCredential `json:"-"`
}

// WebAuthnCredential 存储 Passkeys/FIDO2 密钥
type WebAuthnCredential struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	WebAuthnID      []byte    `json:"webauthn_id"`      // Credential ID (Raw Bytes)
	PublicKey       []byte    `json:"public_key"`       // Public Key (COSE format)
	AttestationType string    `json:"attestation_type"` // e.g., "none", "packed"
	Transport       []string  `json:"transport"`        // ["internal", "usb", ...]
	AAGUID          []byte    `json:"aaguid"`           // Authenticator Attestation GUID
	SignCount       uint32    `json:"sign_count"`       // 防重放计数器
	CloneWarning    bool      `json:"clone_warning"`    // 是否检测到克隆
	DeviceName      string    `json:"device_name"`      // e.g. "iPhone 15 Pro"
	CreatedAt       time.Time `json:"created_at"`
	LastUsedAt      time.Time `json:"last_used_at"`
}

// --- webauthn.User Interface Implementation ---

func (u *User) WebAuthnID() []byte {
	return []byte(u.UUID)
}

func (u *User) WebAuthnName() string {
	return u.Username
}

func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *User) WebAuthnIcon() string {
	return "" // Optional
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, c := range u.Credentials {
		creds[i] = webauthn.Credential{
			ID:              c.WebAuthnID,
			PublicKey:       c.PublicKey,
			AttestationType: c.AttestationType,
			Transport:       c.Transport,
			Authenticator: webauthn.Authenticator{
				AAGUID:       c.AAGUID,
				SignCount:    c.SignCount,
				CloneWarning: c.CloneWarning,
			},
		}
	}
	return creds
}

// Helper: Convert Database JSONB to Transport Slice
func ParseTransports(data []byte) []string {
	var t []string
	_ = json.Unmarshal(data, &t)
	return t
}
