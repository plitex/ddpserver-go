package ddpserver

type PublicationHandler func(ctx SubscriptionContext) (result interface{}, err *Error)

type SubscriptionContext struct {
	ID     string
	Params []interface{}
	conn   Session
	ready  bool
}

func NewSubscriptionContext(m Message, conn Session) *SubscriptionContext {
	ctx := &SubscriptionContext{}
	ctx.conn = conn
	ctx.ID = m.ID
	ctx.Params = m.Params
	return ctx
}

func (ctx *SubscriptionContext) Ready() {
	ctx.ready = true

	ctx.conn.sendSubscriptionReady(ctx)
}

func (ctx *SubscriptionContext) Added(collection string, id string, fields map[string]interface{}) {
	ctx.conn.sendAdded(ctx, collection, id, fields)
}

func (ctx *SubscriptionContext) Changed(collection string, id string, fields map[string]interface{}, cleared map[string]interface{}) {
	ctx.conn.sendChanged(ctx, collection, id, fields, cleared)
}

func (ctx *SubscriptionContext) Removed(collection string, id string) {
	ctx.conn.sendRemoved(ctx, collection, id)
}
