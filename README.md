# Fields and Middleware for REST Layerto implememy TTL

Fields and Middleware for [REST Layer](https://github.com/rs/rest-layer) to implememy TTL (Time-To-Live) on those Storage backends that do not support it.

## Usage

```go
import "github.com/apuigsech/rest-layer-ttl"
```

You need to create three fields on the schema;

```go
    unit = schema.Schema{
        Fields: schema.Fields{
            "id": schema.IDField,
            "created": schema.CreatedField,
            "updated": schema.UpdatedField,

            "ttl": ttl.TTLField,
            "deleteat": ttl.DeleteAtField,
            "active": ttl.ActiveField,
    }

```

Create the Middleware using _NewTTLMiddleWare_ and agregate it into REST Layer with the _Use_ function:

```go
units.Use(ttl.NewTTLMiddleWare("ttl", "deleteat", "active", false, 0, units))
```

If the parameter autoDeleteItems is set to true, items will be evaluated for deletion with the specified frequency. It it's set to false, the field "active" will ve set to false once the TTL expires but the item won't be deleted.
