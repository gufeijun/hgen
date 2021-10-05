package service

type Message struct {
	Name string
	Mems []*Member
}

type Member struct {
	MemType *Type
	MemName string
}
