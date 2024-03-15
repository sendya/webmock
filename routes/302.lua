local router = require("router")

router.handle("/302", function(req, res)
  res:redirect("/hello")
end)