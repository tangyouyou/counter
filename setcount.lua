local function byte2bin(n, bits)
local t = {}
  for i=bits-1,0,-1 do
    t[#t+1] = math.floor(n / 2^i)
    n = n % 2^i
  end
  if t[1] == -1 then
     t[1] = 1
  end
  return t
end

local cacheKey = KEYS[1]
local offset = tonumber(KEYS[2])
local value = tonumber(KEYS[3])
local bitArr = byte2bin(value, 32)

local startOffset = 0
local offsetIndex = 0
if offset then
   startOffset = offset * 32
else
   return false
end

local bitIndex = 1
for i = 1, #bitArr do
    redis.call("setbit", cacheKey, startOffset + offsetIndex, bitArr[i])
    offsetIndex = offsetIndex + 1
end

return true
