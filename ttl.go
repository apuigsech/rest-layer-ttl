package ttl

import (
	"fmt"
	"errors"
	"time"
	"context"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
	"github.com/rs/rest-layer/resource"
)

var (
	TTLField = schema.Field{
		Description: "Time-to-live in seconds",
		Required: true,
		Default: 0,
		Filterable: true,
		Sortable: true,
		Validator: &schema.Integer{},
	}

	DeleteAtField = schema.Field{
		Description: "Moment in wich item will be deleted",
		Required: true,
		Default: time.Unix(1<<63-1, 0).Unix(),
		Filterable: true,
		Sortable: true,
		Validator: &schema.Integer{},
	}

	ActiveField = schema.Field{
		Description: "Is the itema active due TTL?",
		Required: true,
		Default: true,
		Filterable: true,
		Validator: &schema.Bool{},
	}
)

type TTLMiddleWare struct {
	TTLFieldName string
	DeleteAtFieldName string
	ActiveFieldName string
	AutoDeleteItems bool
	resource *resource.Resource
}

func AnyInt(value interface{}) (int_val int, int_ok bool, int32_val int32, int32_ok bool, int64_val int64, int64_ok bool) {
	int_val, int_ok = value.(int)
	int32_val, int32_ok = value.(int32)
	int64_val, int64_ok = value.(int64)
	
	if int_ok {
		int32_ok = true
		int32_val = int32(int_val)
	}

	if int32_ok {
		int64_ok = true
		int64_val = int64(int_val)
	}
	
	return
}


func Int64(value interface{}) (int64_val int64, int64_ok bool) {
	_, _, _, _, int64_val, int64_ok = AnyInt(value)
	return
}


func NewTTLMiddleWare(ttlFieldName string, deleteAtFieldName string, activeFieldName string, autoDeleteItems bool, interval int, rsc *resource.Resource) (TTLMiddleWare) {
	if interval > 0  && autoDeleteItems {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)

		go func() {
			for _ = range ticker.C {
				q, err := query.New("", fmt.Sprintf("{%s: {$lte: \"%d\"}}", deleteAtFieldName, time.Now().Local().Unix()), "", nil)
				if err == nil {
					rsc.Clear(context.TODO(), q)
				}
				// TODO: What to do if error?
			}
		}()
	}

	return TTLMiddleWare{
		TTLFieldName: ttlFieldName,
		DeleteAtFieldName: deleteAtFieldName,
		ActiveFieldName: activeFieldName,
		AutoDeleteItems: autoDeleteItems,
		resource: rsc,
	}
}


func (mw TTLMiddleWare) OnInsert(ctx context.Context, items []*resource.Item) error {
	for _, i := range items {

		ttl, ok := Int64(i.Payload[mw.TTLFieldName])
		if !ok {
			return errors.New("TTLField not found")
		}

		if ttl > 0 {
			i.Payload[mw.DeleteAtFieldName] = time.Now().Local().Add(time.Duration(ttl) * time.Second).Unix()
		}
	}

	return nil
}

func (mw TTLMiddleWare) OnUpdate(ctx context.Context, item *resource.Item, original *resource.Item) error {
	var ttl int64

	if !item.Payload[mw.ActiveFieldName].(bool) {
		return nil
	}

	ttl_item, ok_item := Int64(item.Payload[mw.TTLFieldName])
	ttl_original, ok_original := Int64(original.Payload[mw.TTLFieldName])

	if !ok_item && !ok_original {
		return errors.New("TTLField not found")
	}

	if ok_item {
		ttl = ttl_item
	} else {
		if ok_original {
			ttl = ttl_original
		}
	}

	item.Payload[mw.DeleteAtFieldName] = time.Now().Local().Add(time.Duration(ttl) * time.Second).Unix()

	return nil
}


func (mw TTLMiddleWare) OnFound(ctx context.Context, query *query.Query, list **resource.ItemList, err *error) {
	if !mw.AutoDeleteItems {
		for _, i := range (*list).Items {
			if i.Payload[mw.DeleteAtFieldName].(int64) <= time.Now().Local().Unix() {
				i.Payload[mw.ActiveFieldName] = false
				mw.resource.Update(ctx, i, i)
			}
		}
	}
}

func (mw TTLMiddleWare) OnGot(ctx context.Context, item **resource.Item, err *error) {
	i := *item
	if !mw.AutoDeleteItems {
		if i.Payload[mw.DeleteAtFieldName].(int64) <= time.Now().Local().Unix() {
			i.Payload[mw.ActiveFieldName] = false
			mw.resource.Update(ctx, i, i)
		}
	}
}
