-- get.lua
counter = 0

request = function()
    counter = counter + 1
    local key = "key_" .. (counter % 10000)
    return wrk.format("GET", "/get?key=" .. key)
end