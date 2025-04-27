package verify_code

import "context"

type VerifyCodeOption struct {
	Interval     int `yaml:"interval"`
	Limit        int `yaml:"limit"`
	Lifetime     int `yaml:"lifetime"`
	ErrorTimes   int `yaml:"error_times"`
	SuccessTimes int `yaml:"success_times"`
}

type VerifyCode interface {
	StoreVerifyCode(ctx context.Context, key, code, ip string) (int, error)
	ValidateVerifyCode(ctx context.Context, key, code, ip string) error
	ResetVerifyCode(ctx context.Context, key string) error
}
