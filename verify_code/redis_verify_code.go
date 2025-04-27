package verify_code

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisVerifyCodeSuccess                = 0
	redisStoreVerifyCodeErrFastCode       = 1
	redisStoreVerifyCodeErrLimitCode      = 2
	redisValidateVerifyCodeErrInvalidCode = 3
	redisValidateVerifyCodeErrExpiredCode = 4
	redisNeedResetVerifyCodeErrCode       = 5
)

var (
	RedisStoreVerifyCodeErrFast    = errors.New("verify code is too fast")
	RedisStoreVerifyCodeErrLimit   = errors.New("verify code is too many")
	RedisValidateVerifyCodeErr     = errors.New("verify code is invalid")
	RedisValidateVerifyCodeExpired = errors.New("verify code is expired")
	RedisNeedResetVerifyCode       = errors.New("verify code need reset")
)

var errMap = map[int]error{
	redisVerifyCodeSuccess:                nil,
	redisStoreVerifyCodeErrFastCode:       RedisStoreVerifyCodeErrFast,
	redisStoreVerifyCodeErrLimitCode:      RedisStoreVerifyCodeErrLimit,
	redisValidateVerifyCodeErrInvalidCode: RedisValidateVerifyCodeErr,
	redisValidateVerifyCodeErrExpiredCode: RedisValidateVerifyCodeExpired,
	redisNeedResetVerifyCodeErrCode:       RedisNeedResetVerifyCode,
}

type redisVerifyCodeService struct {
	option      VerifyCodeOption
	rdb         redis.UniversalClient
	keyTemplate string
}

func (s *redisVerifyCodeService) StoreVerifyCode(ctx context.Context, key, code, ip string) (int, error) {
	script := redis.NewScript(redisStoreVerifyCodeScript)
	verifyCodeKey := fmt.Sprintf(s.keyTemplate, key)
	slice, err := script.Run(
		ctx,
		s.rdb,
		[]string{verifyCodeKey},
		ip,
		code,
		time.Now().Unix(),
		s.option.Interval,
		s.option.Limit,
		redisStoreVerifyCodeErrFastCode,
		redisStoreVerifyCodeErrLimitCode,
	).Int64Slice()

	if err != nil {
		return 0, err
	}

	errCode := int(slice[0])
	interval := int(slice[1])

	return interval, errMap[errCode]
}
func (s *redisVerifyCodeService) ValidateVerifyCode(ctx context.Context, key, code, ip string) error {
	script := redis.NewScript(redisValidateVerifyCodeScript)
	verifyCodeKey := fmt.Sprintf(s.keyTemplate, key)
	errCode, err := script.Run(
		ctx,
		s.rdb,
		[]string{verifyCodeKey},
		ip,
		code,
		time.Now().Unix(),
		s.option.Lifetime,
		s.option.ErrorTimes,
		s.option.SuccessTimes,
		redisNeedResetVerifyCodeErrCode,
		redisValidateVerifyCodeErrExpiredCode,
		redisValidateVerifyCodeErrInvalidCode,
	).Int()

	if err != nil {
		return err
	}

	return errMap[errCode]
}

func (s *redisVerifyCodeService) ResetVerifyCode(ctx context.Context, key string) error {
	script := redis.NewScript(redisResetVerifyCodeScript)
	smsVerifyCodeKey := fmt.Sprintf(s.keyTemplate, key)
	_, err := script.Run(ctx, s.rdb, []string{smsVerifyCodeKey}).Int()
	if err != nil {
		return err
	}
	return nil
}

func NewRedisVerifyCodeService(rdb redis.UniversalClient, opt VerifyCodeOption, template string) VerifyCode {
	return &redisVerifyCodeService{
		option:      opt,
		rdb:         rdb,
		keyTemplate: template,
	}
}
