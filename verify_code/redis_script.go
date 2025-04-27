package verify_code

const redisStoreVerifyCodeScript = `
		local key =  KEYS[1]
		local ip = ARGV[1]
		local code = ARGV[2]
		local now = tonumber(ARGV[3])
		local interval = tonumber(ARGV[4])
		local dayTotal = tonumber(ARGV[5])
		local errFast = tonumber(ARGV[6])
		local errLimit = tonumber(ARGV[7])
	
		local time = tonumber(redis.call('HGET', key, 'time'))
		if time and now - time < interval then
			return {errFast,(interval + time - now)}
		end
		local total = tonumber(redis.call('HGET', key, 'total'))
		if total and total >= dayTotal then
			return {errLimit,0}
		end
		redis.call('HMSET', key, 'code', code, 'time', now, 'ip', ip, 'success', 0, 'error', 0)
		redis.call('HINCRBY', key, 'total', 1)
		redis.call('EXPIRE', key, 3600 * 24)
		return {0,interval}
	`
const redisValidateVerifyCodeScript = `
		local key =  KEYS[1]
		local ip = ARGV[1]
		local code = ARGV[2]
		local now = tonumber(ARGV[3])

		local lifetime = tonumber(ARGV[4])
		local errorTimes = tonumber(ARGV[5])
		local successTimes = tonumber(ARGV[6])

		local errReset = tonumber(ARGV[7])
		local errExpired = tonumber(ARGV[8])
		local errInvalid = tonumber(ARGV[9])

		local currCode = redis.call('HGET', key, 'code')
		if type(currCode) == 'string' and string.len(currCode) > 0 then
			if currCode ~= code then
				local curErrorTimes = redis.call('HINCRBY', key, 'error', 1)
				if curErrorTimes > errorTimes then
					redis.call('HSET', key, 'code', '')
				end
				return errInvalid
			end
			local time = tonumber(redis.call('HGET', key, 'time'))
			if not time or now - time > lifetime then
				return errExpired
			end
			local curSuccessTimes = redis.call('HINCRBY', key, 'success', 1)
			if curSuccessTimes > successTimes then
				redis.call('HSET', key, 'code', '')
			end
			return 0
		end
		local retArr = redis.call('HMGET', key, 'success','error')
		local successTime = tonumber(retArr[1])
		local errorTimes = tonumber(retArr[2])
		if ( errorTimes and errorTimes > 0 ) or ( successTime and successTime > 0 ) then
			return errReset
		end    
		return errInvalid
	`
const redisResetVerifyCodeScript = `
		local key = KEYS[1]
		local res = redis.call('HSET', key, 'code', '')
		return res
	`
