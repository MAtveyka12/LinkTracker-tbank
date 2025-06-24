package service

type Service interface {
	ChatService
	LinkService
}

type service struct {
	ChatService
	LinkService
}

func NewService(chatService ChatService, linkService LinkService) Service {
	return &service{
		ChatService: chatService,
		LinkService: linkService,
	}
}
