package config

type AWSSESCOnfig struct {
	Region string `json:"region,optional,env=AWS_REGION"`      // 发送邮件的区域
	Email  string `json:"email,optional,env=AWS_SENDER_EMAIL"` // 发送邮件的邮箱
}
type AWSConfig struct {
	AccessKeyID     string       `json:"access_key_id,optional,env=AWS_ACCESS_KEY_ID"`         // aws access key id
	SecretAccessKey string       `json:"secret_access_key,optional,env=AWS_SECRET_ACCESS_KEY"` // aws secret access key
	SESConfig       AWSSESCOnfig `json:"ses_config"`                                           // aws ses config
}
