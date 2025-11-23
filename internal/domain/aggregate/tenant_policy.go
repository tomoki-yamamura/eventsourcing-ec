package aggregate

import (
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

type TenantPolicyAggregate struct {
	tenantID             uuid.UUID
	cartAbandonedMinutes int
	version              int
	uncommitted          []event.Event
}

func NewTenantPolicyAggregate() *TenantPolicyAggregate {
	return &TenantPolicyAggregate{
		version:     -1,
		uncommitted: make([]event.Event, 0),
	}
}

func (a *TenantPolicyAggregate) GetAggregateID() uuid.UUID { return a.tenantID }
func (a *TenantPolicyAggregate) GetVersion() int           { return a.version }

func (a *TenantPolicyAggregate) GetUncommittedEvents() []event.Event {
	return a.uncommitted
}

func (a *TenantPolicyAggregate) MarkEventsAsCommitted() {
	a.uncommitted = nil
}

func (a *TenantPolicyAggregate) Hydration(events []event.Event) error {
	for _, ev := range events {
		a.apply(ev)
	}
	return nil
}

func (a *TenantPolicyAggregate) apply(ev event.Event) {
	switch ev.(type) {
	// TODO: イベントが実装されたら追加
	// case *event.TenantPolicyCreated:
	// 	a.tenantID = e.GetAggregateID()
	// 	a.cartAbandonedMinutes = e.CartAbandonedMinutes
	// 	a.version = e.GetVersion()
	// case *event.CartAbandonedPolicyChanged:
	// 	a.cartAbandonedMinutes = e.NewMinutes
	// 	a.version = e.GetVersion()
	default:
		// 未知のイベントは無視
	}
}

// CartAbandonedDelay はドメインロジックで使う値
func (a *TenantPolicyAggregate) CartAbandonedDelay() time.Duration {
	return time.Duration(a.cartAbandonedMinutes) * time.Minute
}

// 初回作成（管理画面でテナント作成時など）
func (a *TenantPolicyAggregate) Create(tenantID uuid.UUID, minutes int) error {
	if a.version != -1 {
		return errors.UnpermittedOp.New("tenant policy already exists")
	}
	
	// 直接状態を更新（イベントは後で実装）
	a.tenantID = tenantID
	a.cartAbandonedMinutes = minutes
	a.version = 1
	
	// TODO: イベントが実装されたら追加
	// ev := event.NewTenantPolicyCreated(tenantID, 1, minutes)
	// a.apply(ev)
	// a.uncommitted = append(a.uncommitted, ev)
	
	return nil
}

// 設定変更（管理画面から）
func (a *TenantPolicyAggregate) ChangeCartAbandonedMinutes(newMinutes int) error {
	if a.version == -1 {
		return errors.UnpermittedOp.New("tenant policy not created")
	}
	if a.cartAbandonedMinutes == newMinutes {
		return nil
	}
	
	// 直接状態を更新（イベントは後で実装）
	a.cartAbandonedMinutes = newMinutes
	a.version++
	
	// TODO: イベントが実装されたら追加
	// ev := event.NewCartAbandonedPolicyChanged(a.tenantID, a.version, newMinutes)
	// a.apply(ev)
	// a.uncommitted = append(a.uncommitted, ev)
	
	return nil
}