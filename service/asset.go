package service

type Asset struct {
	Services []*Service
	Messages []*Message
}

var GlobalAsset = &Asset{}

func SaveService(service *Service) {
	GlobalAsset.Services = append(GlobalAsset.Services, service)
}

func SaveMessage(message *Message) {
	GlobalAsset.Messages = append(GlobalAsset.Messages, message)
}
