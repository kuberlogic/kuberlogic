local typedefs = require "kong.db.schema.typedefs"


return {
    name = "keycloak-token-introspect",
    fields = {
        {
            consumer = typedefs.no_consumer,
        },
        {
            protocols = typedefs.protocols_http,
        },
        {
            config = {
                type = "record",
                fields = {
                    -- if a JWT token is required
                    {  token_required = { type     = "boolean", required = true, default  = true }, },
                    -- query arg to search token in
                    {  query_arg = { type    = "string", required = true }, },
                    -- url for introspection
                    {  introspection_url = { type    = "string", required = true }, },
                    -- username for introspection basic auth
                    {  basic_username = { type    = "string", required = true }, },
                    -- password for introspection basic auth
                    {  basic_password = { type    = "string", required = true }, },
                    -- token claim to export in X-INTROSPECTION-ID header
                    {  id_claim = { type    = "string", required = true }, },
                },
            },
        },
    },
}
