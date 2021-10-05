package service

const (
	TypeKindNormal = iota
	TypeKindStream
	TypeKindMessage
)

var BuiltinTypes = map[string]struct{}{
	"int8":    struct{}{},
	"uint8":   struct{}{},
	"int16":   struct{}{},
	"uint16":  struct{}{},
	"int32":   struct{}{},
	"uint32":  struct{}{},
	"int64":   struct{}{},
	"uint64":  struct{}{},
	"float32": struct{}{},
	"float64": struct{}{},
	"string":  struct{}{},
	"stream":  struct{}{},
	"istream": struct{}{},
	"ostream": struct{}{},
	"void":    struct{}{},
}

type Type struct {
	TypeKind   uint16 //0 normal, 1 stream, 2 message
	TypeName   string
	FieldTypes []string
}

func IsBuiltinType(_type string) bool {
	_, ok := BuiltinTypes[_type]
	return ok
}

func IsStreamType(t string) bool {
	return t == "istream" || t == "ostream" || t == "stream"
}
