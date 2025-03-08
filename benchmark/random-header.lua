math.randomseed(os.time())

request = function()
    local id = math.random(0, 1000)
    local headers = {["X-ID"] = tostring(id)}
    return wrk.format(nil, nil, headers, nil)
end
