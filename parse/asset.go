package parse

const (
	TypeKindNormal = iota
	TypeKindStream
	TypeKindMessage
	TypeKindErr
	TypeKindNoRTN
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

type Symbols struct {
	Services map[string]*Service
	Messages map[string]*Message
}

type Service struct {
	Name    string    // 服务名
	Methods []*Method // 这个服务下的所有方法
}

type Message struct {
	Name string    // Message名
	Mems []*Member // 包含的成员
}

type Method struct {
	Service  *Service // 这是属于哪个服务的方法
	RetType  *Type    // 方法返回值
	ReqTypes []*Type  // 方法请求参数
	Name     string   // 方法名
}

type Member struct {
	MemType *Type  // 成员的类型信息
	MemName string // 成员名
}

type Type struct {
	Kind uint16 // 0 normal, 1 stream, 2 message
	Name string // 类型名
}

func newType(name string) *Type {
	t := &Type{Name: name}
	if _, ok := BuiltinTypes[name]; !ok {
		t.Kind = TypeKindMessage
	} else if name == "stream" || name == "istream" || name == "ostream" {
		t.Kind = TypeKindStream
	} else {
		t.Kind = TypeKindNormal
	}
	return t
}
