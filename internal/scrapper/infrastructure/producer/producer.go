package producer

import prod "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"

type ProdKafkaOrHTTP interface {
	Run()
	Stop()
	SendUpdate(update prod.LinkUpdate) error
	SendUpdatePR(update prod.LinkUpdate) error
}
