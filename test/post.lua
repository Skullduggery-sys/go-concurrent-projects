-- post.lua
counter = 0

request = function()
    counter = counter + 1
    local body = string.format('{"key":"key_%d","value":"value_%d","ttl":600000000000}', counter, counter)
    local headers = {["Content-Type"] = "application/json"}
    return wrk.format("POST", "/set", headers, body)
end