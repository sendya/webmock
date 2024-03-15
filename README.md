webmock
---


### Usage

```go
$ go run main.go
```

### New Handler

```lua
local router = require('router')

router.handle("/path/to", function(req, res)
  -- show host
  print(req.host)

  res.header("X-Test-Name", "aaabbcc")
  res.write('OK')
end)

router.handle("/4xx", function(req, res)
  res.header("X-Test-Name", "required Authorization")
  res.write_header(401)
  res.write('')
end)
```
