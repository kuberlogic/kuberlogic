local httpc = require "resty.http"
local encode_base64 = ngx.encode_base64
local JSON = require "kong.plugins.keycloak-token-introspect.JSON"

local KeycloakTokenIntrospect = {
    PRIORITY = 900,
    VERSION = "1.0"
}

local function do_auth(conf)
    local token = kong.request.get_query_arg(conf.query_arg)
    if token ~= nil then
        local http_connection = httpc:new()
        local r, err = http_connection:request_uri(conf.introspection_url, {
            method = "POST",
            ssl_verify = false,
            body = "token_type_hint=access_token&token="..token,
            headers = {
                ["Content-Type"] = "application/x-www-form-urlencoded",
                ["Authorization"] = "Basic "..encode_base64(conf.basic_username..":"..conf.basic_password),
            }
        })

        if err ~= nil then
            return "", "keycloak introspection error "..err
        end
        if r.status ~= 200 then
            return "", "keycloak introspection unknown status "..r.status
        end

        local data = JSON:decode(r.body)
        if data == nil then
            return "", "keycloak introspection decode failure"
        end
        if data["active"] ~= true then
            return "", ""
        end
        local id_claim_v = data[conf.id_claim]
        if id_claim_v == nil then
            return "", "Failed to find an id claim in response "..conf.id_claim
        end

        return id_claim_v, nil
    end
    return "", ""
end

function KeycloakTokenIntrospect:access(conf)
    if not conf.token_required then
        return
    end
    local id, err = do_auth(conf)
    if err ~= nil then
        kong.log.err(err)
        kong.response.exit(501, "error verifying authentication token")
    end
    if id == "" then
        kong.response.exit(401, "token is not active")
    end
    kong.log.debug("authenticated "..id)

    kong.service.request.set_header("X-INTROSPECTION-ID", id)
end

return KeycloakTokenIntrospect
