package service

type Method struct {
	Service    *Service
	RetType    *Type
	ReqTypes   []*Type
	MethodName string
}
