package pubsub

import (
	"ElectronicQueue/internal/logger"
	"sync"
)

type Broker struct {
	mu          sync.Mutex
	subscribers map[chan string]bool
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[chan string]bool),
	}
}

// Subscribe добавляет нового подписчика (клиента).
// Возвращает канал, который будет получать сообщения.
func (b *Broker) Subscribe() chan string {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan string, 10) // Буферизированный канал, чтобы не блокировать рассылку
	b.subscribers[ch] = true
	logger.Default().Info("PubSub: New client subscribed.")
	return ch
}

// Unsubscribe удаляет подписчика.
func (b *Broker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
		logger.Default().Info("PubSub: Client unsubscribed.")
	}
}

// Publish отправляет сообщение всем активным подписчикам.
func (b *Broker) Publish(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	log := logger.Default()
	log.WithField("subscribers", len(b.subscribers)).Info("PubSub: Publishing message.")

	for ch := range b.subscribers {
		// Используем неблокирующую отправку, чтобы один "медленный" клиент
		// не затормозил рассылку для всех остальных.
		select {
		case ch <- msg:
		default:
			log.Warn("PubSub: Message channel for a client is full. Message dropped.")
		}
	}
}

// ListenAndPublish - это горутина, которая слушает входящий канал
// от PostgreSQL и публикует сообщения через брокер.
func (b *Broker) ListenAndPublish(notifications <-chan string) {
	for msg := range notifications {
		b.Publish(msg)
	}
}
