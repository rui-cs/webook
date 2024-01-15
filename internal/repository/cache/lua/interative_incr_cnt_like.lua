local key = KEYS[1]
-- 对应到的是 hincrby 中的 field
local cntKey = ARGV[1]
-- +1 或者 -1
local delta = tonumber(ARGV[2])
local likeKey = ARGV[3]
local biz = ARGV[4]
local bizID = ARGV[5]
local threshold = tonumber(ARGV[6])
local exists = redis.call("EXISTS", key)
if exists == 1 then
    redis.call("HINCRBY", key, cntKey, delta)
    -- 说明自增成功了

    -- 取出数据
    local likeCnt = tonumber(redis.call("HGET", key, likeKey))
    -- 如果大于10w，加入到hotlist的key中
    --if likeCnt >= 100000 then
    if likeCnt >= threshold then
        --    zadd hotlist:biz:video:like 22 id1
        local hotlistKey = "hotlist:biz:" .. biz .. ":like"
        redis.call("ZADD", hotlistKey, likeCnt, bizID)
    end

    return 1
else
    -- 自增不成功
    return 0
end