package ddpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeField(t *testing.T) {
	ctx1 := &SubscriptionContext{}
	d := NewSessionDocumentView()
	changeCollector := NewSessionDocumentChangeCollector()

	// Change field value from subscription 1
	assert.Equal(t, true, d.IsEmpty())
	d.ChangeField(ctx1, "field", "Value1", changeCollector, true)
	assert.Equal(t, 1, len(d.existsIn))
	assert.Equal(t, true, d.existsIn[ctx1])
	assert.Equal(t, 1, len(d.dataByKey))
	assert.Equal(t, 1, len(d.dataByKey["field"]))
	assert.Equal(t, "Value1", d.dataByKey["field"][0].value)
	assert.Equal(t, 1, len(changeCollector.changed))
	assert.Equal(t, 0, len(changeCollector.removed))
	assert.Equal(t, "Value1", changeCollector.changed["field"])
	assert.Equal(t, false, d.IsEmpty())

	// Change field value from subscription 2
	ctx2 := &SubscriptionContext{}
	changeCollector = NewSessionDocumentChangeCollector()
	d.ChangeField(ctx2, "field", "Value1", changeCollector, false)
	assert.Equal(t, 2, len(d.existsIn))
	assert.Equal(t, true, d.existsIn[ctx1])
	assert.Equal(t, true, d.existsIn[ctx2])
	assert.Equal(t, 1, len(d.dataByKey))
	assert.Equal(t, 2, len(d.dataByKey["field"]))
	assert.Equal(t, "Value1", d.dataByKey["field"][1].value)
	assert.Equal(t, 0, len(changeCollector.changed))
	assert.Equal(t, 0, len(changeCollector.removed))

	// Change field value again subscription 1
	changeCollector = NewSessionDocumentChangeCollector()
	d.ChangeField(ctx1, "field", "Value2", changeCollector, false)
	assert.Equal(t, 2, len(d.existsIn))
	assert.Equal(t, true, d.existsIn[ctx1])
	assert.Equal(t, true, d.existsIn[ctx2])
	assert.Equal(t, 1, len(d.dataByKey))
	assert.Equal(t, 2, len(d.dataByKey["field"]))
	assert.Equal(t, "Value2", d.dataByKey["field"][0].value)
	assert.Equal(t, 1, len(changeCollector.changed))
	assert.Equal(t, 0, len(changeCollector.removed))
	assert.Equal(t, "Value2", changeCollector.changed["field"])

	// Change field value again subscription 2
	changeCollector = NewSessionDocumentChangeCollector()
	d.ChangeField(ctx2, "field", "Value2", changeCollector, false)
	assert.Equal(t, 2, len(d.existsIn))
	assert.Equal(t, true, d.existsIn[ctx1])
	assert.Equal(t, true, d.existsIn[ctx2])
	assert.Equal(t, 1, len(d.dataByKey))
	assert.Equal(t, 2, len(d.dataByKey["field"]))
	assert.Equal(t, "Value2", d.dataByKey["field"][1].value)
	assert.Equal(t, 0, len(changeCollector.changed))
	assert.Equal(t, 0, len(changeCollector.removed))

	// Clear field value from subscription 2
	changeCollector = NewSessionDocumentChangeCollector()
	d.ClearField(ctx2, "field", changeCollector)
	assert.Equal(t, 1, len(d.existsIn))
	assert.Equal(t, true, d.existsIn[ctx1])
	assert.Equal(t, false, d.existsIn[ctx2])
	assert.Equal(t, 1, len(d.dataByKey))
	assert.Equal(t, 1, len(d.dataByKey["field"]))
	assert.Equal(t, "Value2", d.dataByKey["field"][0].value)
	assert.Equal(t, 0, len(changeCollector.changed))
	assert.Equal(t, 0, len(changeCollector.removed))

	// Clear field value from subscription 1
	changeCollector = NewSessionDocumentChangeCollector()
	d.ClearField(ctx1, "field", changeCollector)
	assert.Equal(t, 0, len(d.existsIn))
	assert.Equal(t, false, d.existsIn[ctx1])
	assert.Equal(t, false, d.existsIn[ctx2])
	assert.Equal(t, 0, len(d.dataByKey))
	assert.Equal(t, 0, len(d.dataByKey["field"]))
	assert.Equal(t, 0, len(changeCollector.changed))
	assert.Equal(t, 1, len(changeCollector.removed))
	assert.Equal(t, true, d.IsEmpty())
}

func TestGetFields(t *testing.T) {
	d := NewSessionDocumentView()
	ctx := &SubscriptionContext{}
	d.dataByKey["field"] = []*SessionDocumentViewValue{&SessionDocumentViewValue{ctx, "Value"}}

	ret := d.GetFields()
	assert.Equal(t, 1, len(ret))
	assert.Equal(t, "Value", ret["field"])
}
