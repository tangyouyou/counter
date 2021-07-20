local function covertBinToBcd(t)
  local num = 0
  local sign = 1
  for i = 1, #t do
     t[i] = tonumber(t[i])
  end
  if t[1] == 1 then
    sign = -1
    for i=1, #t do
            if t[i] == 1 then
                t[i] = 0
            else
                t[i] = 1
             end
    end
  end
  for i = #t-1,0,-1 do
    local tmp = t[#t-i]
    tmp = tmp == 1 and 1 or 0
    num = num + tmp * math.pow(2, i)
  end
  if sign == -1 then
      num = 0 - (num + 1)
  end
  return num
end

local cacheKey = KEYS[1]
local offset = tonumber(KEYS[2])
local bitArr = {}
local startOffset = 0
if offset then
   startOffset = offset * 32
else
   bitArr[1] = KEYS[1]
   bitArr[2] = KEYS[2]
   return bitArr
end

local bitIndex = 1
for i = startOffset,startOffset + 31 do
    bitArr[bitIndex] = tostring(redis.call("getbit", cacheKey, i))
    bitIndex = bitIndex + 1
end

return covertBinToBcd(bitArr)
