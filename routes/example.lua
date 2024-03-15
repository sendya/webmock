local router = require("router")

router.handle("/example", function(req, res)
  local body, err = req.body()
  if err then error(err) end
  print("body:", body)

  res:header("Content-Type", "application/json")
  res:write_header(400)

  res:json('{"a": 1, "b": '..body..'}')
end)

router.handle("/hello", function(req, res)
  res:header("Vary", "User-Agent, Accept-Encoding")
  res:write('hello world!!!!')
end)