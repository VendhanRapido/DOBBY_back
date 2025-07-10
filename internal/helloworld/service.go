package helloworld

type Service interface {
	HelloWorld() string
}

type serviceImpl struct {
}

func NewService() Service {
	return &serviceImpl{}
}

func (s *serviceImpl) HelloWorld() string {
	return "Hello World!"
}
