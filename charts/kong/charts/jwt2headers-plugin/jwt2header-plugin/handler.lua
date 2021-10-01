-- dynamic routing based on JWT Claim
local type = type
local pairs = pairs

local jwt_decoder = require "kong.plugins.jwt.jwt_parser"


local JWT2Header = {
  PRIORITY = 900,
  VERSION = "1.0"
}

function JWT2Header:access(conf)
    kong.service.request.set_header("X-Kong-JWT-Kong-Proceed", "no")
    kong.log.debug(kong.request.get_query_arg(conf.query_arg))
    local claims
    local header
    if kong.request.get_query_arg(conf.query_arg) ~= nil then
        local jwt, err = jwt_decoder:new(kong.request.get_query_arg(conf.query_arg))
        if err and conf.token_required then
            return false, { status = 401, message = "Bad token; " .. tostring(err) }
        end

        claims = jwt.claims
        header = jwt.header

        for claim, value in pairs(claims) do
            if type(claim) == "string" and type(value) == "string" then
                kong.service.request.set_header("X-Kong-JWT-Claim-" .. claim, value)
                kong.log.debug("Added " .. "X-Kong-JWT-Claim-" .. claim .. ": " .. value)
            end
        end
    else
        if conf.token_required then
            return false, { status = 401, message = "JWT token not found" }
        end
    end
end

return JWT2Header
