local typedefs = require "kong.db.schema.typedefs"


return {
  name = "jwt-to-header",
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
            {  token_required = { type     = "boolean", required = true, default  = true }, },
            {  query_arg = { type    = "string", required = true, default = "jwt" }, },
        },
      },
    },
  },
}
