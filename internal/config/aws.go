package config

type AWSSESCOnfig struct {
	Region string `json:"region"` // 发送邮件的区域
	Email  string `json:"email"`  // 发送邮件的邮箱
}
type AWSConfig struct {
	AccessKeyID     string       `json:"access_key_id"`     // aws access key id
	SecretAccessKey string       `json:"secret_access_key"` // aws secret access key
	SESConfig       AWSSESCOnfig `json:"ses_config"`        // aws ses config
}
