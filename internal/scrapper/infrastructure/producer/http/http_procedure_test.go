package http_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	prod "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/producer/mock"
)

func TestSendUpdate(t *testing.T) {
	update := prod.LinkUpdate{URL: "https://example.com", Description: "Update"}

	t.Run("успешная отправка", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProducer := mock.NewMockProdKafkaOrHTTP(ctrl)
		mockProducer.EXPECT().
			SendUpdate(update).
			Return(nil)

		err := mockProducer.SendUpdate(update)
		require.NoError(t, err)
	})

	t.Run("ошибка сериализации JSON", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProducer := mock.NewMockProdKafkaOrHTTP(ctrl)
		brokenUpdate := prod.LinkUpdate{URL: string([]byte{0xFF, 0xFE, 0xFD})}

		mockProducer.EXPECT().
			SendUpdate(brokenUpdate).
			Return(fmt.Errorf("ошибка сериализации данных"))

		err := mockProducer.SendUpdate(brokenUpdate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка сериализации данных")
	})

	t.Run("ошибка выполнения запроса", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProducer := mock.NewMockProdKafkaOrHTTP(ctrl)
		mockProducer.EXPECT().
			SendUpdate(update).
			Return(fmt.Errorf("ошибка выполнения запроса"))

		err := mockProducer.SendUpdate(update)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка выполнения запроса")
	})

	t.Run("ошибка статус-кода", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProducer := mock.NewMockProdKafkaOrHTTP(ctrl)
		mockProducer.EXPECT().
			SendUpdate(update).
			Return(fmt.Errorf("ошибка отправки уведомления"))

		err := mockProducer.SendUpdate(update)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка отправки уведомления")
	})
}

func TestSendUpdatePR(t *testing.T) {
	update := prod.LinkUpdate{ID: 42, Description: "open"}

	t.Run("успешная отправка", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProducer := mock.NewMockProdKafkaOrHTTP(ctrl)
		mockProducer.EXPECT().
			SendUpdatePR(update).
			Return(nil)

		err := mockProducer.SendUpdatePR(update)
		require.NoError(t, err)
	})

	t.Run("ошибка сериализации JSON", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProducer := mock.NewMockProdKafkaOrHTTP(ctrl)
		brokenUpdate := prod.LinkUpdate{Description: string([]byte{0xFF, 0xFE, 0xFD})}

		mockProducer.EXPECT().
			SendUpdatePR(brokenUpdate).
			Return(fmt.Errorf("ошибка сериализации данных"))

		err := mockProducer.SendUpdatePR(brokenUpdate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка сериализации данных")
	})

	t.Run("ошибка выполнения запроса", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProducer := mock.NewMockProdKafkaOrHTTP(ctrl)
		mockProducer.EXPECT().
			SendUpdatePR(update).
			Return(fmt.Errorf("ошибка выполнения запроса"))

		err := mockProducer.SendUpdatePR(update)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка выполнения запроса")
	})

	t.Run("ошибка статус-кода", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProducer := mock.NewMockProdKafkaOrHTTP(ctrl)
		mockProducer.EXPECT().
			SendUpdatePR(update).
			Return(fmt.Errorf("ошибка отправки уведомления"))

		err := mockProducer.SendUpdatePR(update)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка отправки уведомления")
	})
}
