package parse

// import (
// 	"fmt"
// 	"gufeijun/hustgen/service"
// )

// type checker struct {
// 	messageMap map[string]*service.Message
// 	serviceMap map[string]*service.Service
// }

// func newChecker() *checker {
// 	return &checker{
// 		messageMap: make(map[string]*service.Message),
// 		serviceMap: make(map[string]*service.Service),
// 	}
// }

// func (c *checker) check() error {
// 	if err := c.checkMessages(); err != nil {
// 		return err
// 	}
// 	return c.checkServices()
// }

// func (c *checker) checkServices() error {
// 	services := service.GlobalAsset.Services
// 	for _, s := range services {
// 		if _, ok := c.serviceMap[s.Name]; ok {
// 			return fmt.Errorf("repeatedly defined service: %s", s.Name)
// 		}
// 		if err := c.checkService(s); err != nil {
// 			return err
// 		}
// 		c.serviceMap[s.Name] = s
// 	}
// 	return nil
// }

// func (c *checker) checkService(s *service.Service) error {
// 	methodMap := make(map[string]*service.Method)
// 	for _, method := range s.Methods {
// 		if _, ok := methodMap[method.MethodName]; ok {
// 			return fmt.Errorf("repeatedly defined method: %s of service: %s", method.MethodName, s.Name)
// 		}
// 		if err := c.checkMethod(method); err != nil {
// 			return err
// 		}
// 		methodMap[method.MethodName] = method
// 	}
// 	return nil
// }

// func (c *checker) checkMethod(method *service.Method) error {
// 	t := method.RetType.TypeName
// 	if !c.isTypeValidate(t) {
// 		return fmt.Errorf("undefined type: %s in method: %s of service: %s", t, method.MethodName, method.Service.Name)
// 	}
// 	var streaming bool
// 	for _, tp := range method.ReqTypes {
// 		tp.TypeKind = GetTypeKind(tp.TypeName)
// 		if tp.TypeName == "void" {
// 			method.ReqTypes = nil
// 			return nil
// 		}
// 		if tp.TypeKind == service.TypeKindStream {
// 			if streaming {
// 				return fmt.Errorf("[%s.%s]: method must have at most one stream type in parameters", method.Service.Name, method.MethodName)
// 			}
// 			streaming = true
// 		}
// 		if !c.isTypeValidate(tp.TypeName) {
// 			return fmt.Errorf("undefined type: %s in method: %s of service: %s", tp.TypeName, method.MethodName, method.Service.Name)
// 		}
// 	}
// 	return nil
// }

// func (c *checker) isTypeValidate(typeName string) bool {
// 	_, ok := c.messageMap[typeName]
// 	return ok || service.IsBuiltinType(typeName)
// }

// func (c *checker) checkMessages() error {
// 	messages := service.GlobalAsset.Messages
// 	for _, message := range messages {
// 		if _, ok := c.messageMap[message.Name]; ok {
// 			return fmt.Errorf("repeatedly defined message: %s", message.Name)
// 		}
// 		c.messageMap[message.Name] = message
// 	}
// 	for _, message := range messages {
// 		if err := c.checkMessage(message); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (c *checker) checkMessage(message *service.Message) error {
// 	memMap := make(map[string]*service.Member)
// 	for _, member := range message.Mems {
// 		if _, ok := memMap[member.MemName]; ok {
// 			return fmt.Errorf("repeatedly defined member: %s of message: %s", member.MemName, message.Name)
// 		}
// 		t := member.MemType.TypeName
// 		if !c.isTypeValidate(t) {
// 			return fmt.Errorf("undefined type: %s in message: %s ", t, message.Name)
// 		}
// 		if service.IsStreamType(t) || t == "void" {
// 			return fmt.Errorf("%s can not be used as a member of message", t)
// 		}
// 		memMap[member.MemName] = member
// 	}
// 	return nil
// }
