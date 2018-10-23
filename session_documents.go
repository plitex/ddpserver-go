package ddpserver

type SessionDocumentViewValue struct {
	ctx   *SubscriptionContext
	value interface{}
}

type SessionDocumentChangeCollector struct {
	changed map[string]interface{}
	removed map[string]bool
}

func NewSessionDocumentChangeCollector() *SessionDocumentChangeCollector {
	return &SessionDocumentChangeCollector{
		changed: make(map[string]interface{}),
		removed: make(map[string]bool),
	}
}

type SessionDocumentView struct {
	existsIn  map[*SubscriptionContext]bool
	dataByKey map[string][]*SessionDocumentViewValue
}

func NewSessionDocumentView() *SessionDocumentView {
	return &SessionDocumentView{
		existsIn:  make(map[*SubscriptionContext]bool),
		dataByKey: make(map[string][]*SessionDocumentViewValue),
	}
}

func (s *SessionDocumentView) IsEmpty() bool {
	return len(s.existsIn) == 0
}

func (s *SessionDocumentView) GetFields() map[string]interface{} {
	ret := make(map[string]interface{})
	for key, precedenceList := range s.dataByKey {
		ret[key] = precedenceList[0].value
	}
	return ret
}

func (s *SessionDocumentView) ClearField(ctx *SubscriptionContext, key string, changeCollector *SessionDocumentChangeCollector) {
	// It's okay to clear fields that didn't exist. No need to throw
	// an error.
	if _, ok := s.dataByKey[key]; !ok {
		return
	}

	precedenceList := s.dataByKey[key]
	hasRemovedValue := false
	var removedValue interface{}
	for i, precedence := range precedenceList {
		if precedence.ctx == ctx {
			// The view's value can only change if this subscription is the one that
			// used to have precedence.
			if i == 0 {
				removedValue = precedence.value
				hasRemovedValue = true
			}
			precedenceList = append(precedenceList[:i], precedenceList[i+1:]...)
			s.dataByKey[key] = precedenceList
			delete(s.existsIn, ctx)
			break
		}
	}

	if len(precedenceList) == 0 {
		delete(s.dataByKey, key)
		changeCollector.removed[key] = true
	} else if hasRemovedValue && removedValue != precedenceList[0].value {
		changeCollector.changed[key] = precedenceList[0].value
	}
}

func (s *SessionDocumentView) ChangeField(ctx *SubscriptionContext, key string, value interface{},
	changeCollector *SessionDocumentChangeCollector, isAdd bool) {

	// Don't share state with the data passed in by the user.
	// valueType := reflect.TypeOf(value)
	// switch valueType.Kind() {
	// case reflect.String:
	// case reflect.Bool:
	// case reflect.Int:
	// case reflect.Int8:
	// case reflect.Int16:
	// case reflect.Int32:
	// case reflect.Int64:
	// case reflect.Uint:
	// case reflect.Uint8:
	// case reflect.Uint16:
	// case reflect.Uint32:
	// case reflect.Uint64:
	// 	return "ByValue"
	// }

	if _, ok := s.dataByKey[key]; !ok {
		slice := []*SessionDocumentViewValue{}
		slice = append(slice, &SessionDocumentViewValue{ctx, value})
		s.dataByKey[key] = slice
		s.existsIn[ctx] = true
		changeCollector.changed[key] = value
		return
	}

	precedenceList := s.dataByKey[key]
	index := -1
	if !isAdd {
		for i, v := range precedenceList {
			if v.ctx == ctx {
				index = i
			}
		}
	}

	if index >= 0 {
		if index == 0 {
			// this subscription is changing the value of this field.
			changeCollector.changed[key] = value
		}
		elt := precedenceList[index]
		elt.value = value
	} else {
		// this subscription is newly caring about this field
		precedenceList = append(precedenceList, &SessionDocumentViewValue{ctx, value})
		s.dataByKey[key] = precedenceList
		s.existsIn[ctx] = true
	}
}
