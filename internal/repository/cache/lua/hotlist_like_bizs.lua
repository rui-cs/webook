local bizs = KEYS
local res = {}
for i = 1, #bizs do
    local curBiz = "hotlist:biz:" .. bizs[i] .. ":like"
    local bizHotList = redis.call("zrange", curBiz,"0" ,"99", "rev", "withscores")
    -- 添加新元素到数组末尾
    res[#res + 1] = bizHotList
end

return res
