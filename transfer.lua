function covertBinToBcd(t)
num = 0
  sign = 1
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
  for i=#t-1,0,-1 do
    tmp = t[#t-i]
    tmp = tmp == 1 and 1 or 0
    num = num + tmp * math.pow(2, i)
  end
  if sign == -1 then
      num = 0 - (num + 1)
  end
  return num
end

function byte2bin(n, bits)
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

a = table.concat(byte2bin(2147483647,32))
print(a)

b = table.concat(byte2bin(127,32))
print(b)

c = table.concat(byte2bin(-127,32))
print(c)

d = table.concat(byte2bin(5,32))
print(d)

e = table.concat(byte2bin(-5,32))
print(e)

f = table.concat(byte2bin(0,32))
print(f)

h = table.concat(byte2bin(-2147483648,32))
print(h)


t1 = {0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1}
a1 = covertBinToBcd(t1)
print(a1)

t2 = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,1,1,1}
b1 = covertBinToBcd(t2)
print(b1)

t3 = {1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,0,0,1}
c1 = covertBinToBcd(t3)
print(c1)

t4 = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1}
d1 = covertBinToBcd(t4)
print(d1)

t5 = {1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1}
e1 = covertBinToBcd(t5)
print(e1)

t5 = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}
f1 = covertBinToBcd(t5)
print(f1)

t6 = {1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}
h1 = covertBinToBcd(t6)
print(h1)