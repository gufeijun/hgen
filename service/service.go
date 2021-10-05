package service

type Service struct {
	Name    string
	Methods []*Method
}

func NewService(name string) *Service {
	return &Service{
		Name: name,
	}
}

func (s *Service) AddMethod(method *Method) {
	method.Service = s
	s.Methods = append(s.Methods, method)
}
