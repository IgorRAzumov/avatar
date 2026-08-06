package observability

import "go.opentelemetry.io/otel/attribute"

type Attr struct {
	key   string
	value any
}

func String(key, value string) Attr {
	return Attr{key: key, value: value}
}

func Int(key string, value int) Attr {
	return Attr{key: key, value: value}
}

func Int64(key string, value int64) Attr {
	return Attr{key: key, value: value}
}

func (attr Attr) otel() attribute.KeyValue {
	switch value := attr.value.(type) {
	case string:
		return attribute.String(attr.key, value)
	case int:
		return attribute.Int(attr.key, value)
	case int64:
		return attribute.Int64(attr.key, value)
	default:
		return attribute.String(attr.key, "")
	}
}

func attrsToOTel(attrs []Attr) []attribute.KeyValue {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]attribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		out[i] = attr.otel()
	}
	return out
}
