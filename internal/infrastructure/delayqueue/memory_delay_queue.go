package delayqueue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging/dto"
)

type DelayedMessage struct {
	Message   *dto.Message
	Topic     string
	Key       string
	ExecuteAt time.Time
}

type MemoryDelayQueue struct {
	messages map[string]*DelayedMessage
	mu       sync.RWMutex
}

func NewMemoryDelayQueue() *MemoryDelayQueue {
	return &MemoryDelayQueue{
		messages: make(map[string]*DelayedMessage),
	}
}

func (q *MemoryDelayQueue) PublishDelayedMessage(topic, key string, message *dto.Message, delay time.Duration) error {
	executeAt := time.Now().Add(delay)

	delayedMsg := &DelayedMessage{
		Message:   message,
		Topic:     topic,
		Key:       key,
		ExecuteAt: executeAt,
	}

	q.mu.Lock()
	q.messages[message.ID.String()] = delayedMsg
	q.mu.Unlock()

	log.Printf("Scheduled delayed message %s for topic %s, will execute at %s (in %v)",
		message.ID.String(), topic, executeAt.Format(time.RFC3339), delay)

	return nil
}

func (q *MemoryDelayQueue) Start(ctx context.Context) error {
	log.Println("Starting Memory Delay Queue...")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Memory Delay Queue stopped")
			return nil
		case <-ticker.C:
			q.processExpiredMessages()
		}
	}
}

func (q *MemoryDelayQueue) processExpiredMessages() {
	now := time.Now()
	var expiredMessages []*DelayedMessage

	q.mu.Lock()
	for id, delayedMsg := range q.messages {
		if now.After(delayedMsg.ExecuteAt) || now.Equal(delayedMsg.ExecuteAt) {
			expiredMessages = append(expiredMessages, delayedMsg)
			delete(q.messages, id)
		}
	}
	q.mu.Unlock()

	for _, delayedMsg := range expiredMessages {
		q.processMessage(delayedMsg.Topic, delayedMsg.Key, delayedMsg.Message)
	}
}

func (q *MemoryDelayQueue) processMessage(topic, key string, message *dto.Message) {
	// 簡単な処理：ログ出力のみ
	log.Printf(" PROCESSING DELAYED MESSAGE:")
	log.Printf("   Topic: %s", topic)
	log.Printf("   Key: %s", key)
	log.Printf("   Message ID: %s", message.ID.String())
	log.Printf("   Message Type: %s", message.Type)
	log.Printf("   Message Data: %+v", message.Data)
	log.Printf("   Aggregate ID: %s", message.AggregateID.String())
	log.Printf("   Version: %d", message.Version)

	// ここで実際のカゴ落ち通知処理を行う予定
	if message.Type == "CheckCartAbandonmentCommand" {
		if cartID, ok := message.Data.(map[string]any)["cart_id"].(string); ok {
			log.Printf("📧 CART ABANDONMENT DETECTED for cart: %s", cartID)
			log.Printf("   -> Sending abandonment notification email/push...")
		}
	}
}
