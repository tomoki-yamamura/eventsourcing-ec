package aggregate

import (
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

type TenantCartAbandonedPolicyAggregate struct {
	tenantID             uuid.UUID
	title                string
	cartAbandonedMinutes int
	quietTimeFrom        time.Time
	quietTimeTo          time.Time
	version              int
	uncommitted          []event.Event
}


func NewTenantCartAbandonedPolicyAggregate() *TenantCartAbandonedPolicyAggregate {
	return &TenantCartAbandonedPolicyAggregate{
		version:     -1,
		uncommitted: make([]event.Event, 0),
	}
}

func (a *TenantCartAbandonedPolicyAggregate) GetAggregateID() uuid.UUID { return a.tenantID }
func (a *TenantCartAbandonedPolicyAggregate) GetVersion() int           { return a.version }
func (a *TenantCartAbandonedPolicyAggregate) GetTitle() string          { return a.title }

func (a *TenantCartAbandonedPolicyAggregate) GetUncommittedEvents() []event.Event {
	return a.uncommitted
}

func (a *TenantCartAbandonedPolicyAggregate) MarkEventsAsCommitted() {
	a.uncommitted = nil
}

func (a *TenantCartAbandonedPolicyAggregate) Hydration(events []event.Event) error {
	for _, ev := range events {
		a.apply(ev)
	}
	return nil
}

func (a *TenantCartAbandonedPolicyAggregate) apply(ev event.Event) {
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
func (a *TenantCartAbandonedPolicyAggregate) CartAbandonedDelay() time.Duration {
	return time.Duration(a.cartAbandonedMinutes) * time.Minute
}

// IsWithinQuietTime 現在時刻が配信停止時間帯かどうかを判定
func (a *TenantCartAbandonedPolicyAggregate) IsWithinQuietTime(now time.Time) (bool, error) {
	if a.quietTimeFrom.IsZero() || a.quietTimeTo.IsZero() {
		return false, nil
	}

	// 現在時刻とquietTimeをUTCで比較
	nowUTC := now.UTC()
	fromUTC := a.quietTimeFrom.UTC()
	toUTC := a.quietTimeTo.UTC()

	// 時間のみを比較するため、同じ日付で時刻のみを抽出
	currentMinutes := nowUTC.Hour()*60 + nowUTC.Minute()
	fromMinutes := fromUTC.Hour()*60 + fromUTC.Minute()
	toMinutes := toUTC.Hour()*60 + toUTC.Minute()

	if fromMinutes < toMinutes {
		// 例: 01:00〜05:00
		return currentMinutes >= fromMinutes && currentMinutes < toMinutes, nil
	} else if fromMinutes > toMinutes {
		// 例: 22:00〜08:00（またぎパターン）
		return currentMinutes >= fromMinutes || currentMinutes < toMinutes, nil
	} else {
		// from == to の場合は無効扱い
		return false, nil
	}
}

// 初回作成（管理画面でテナント作成時など）
func (a *TenantCartAbandonedPolicyAggregate) Create(cmd command.CreateTenantCartAbandonedPolicyCommand) error {
	if a.version != -1 {
		return errors.UnpermittedOp.New("tenant policy already exists")
	}

	// 直接状態を更新（イベントは後で実装）
	a.tenantID = cmd.TenantID
	a.title = cmd.Title
	a.cartAbandonedMinutes = cmd.AbandonedMinutes
	a.quietTimeFrom = cmd.QuietTimeFrom
	a.quietTimeTo = cmd.QuietTimeTo
	a.version = 1

	// TODO: イベントが実装されたら追加
	// ev := event.NewTenantPolicyCreated(tenantID, 1, title, minutes, quietFrom, quietTo, timezone)
	// a.apply(ev)
	// a.uncommitted = append(a.uncommitted, ev)

	return nil
}

// 設定変更（管理画面から）
func (a *TenantCartAbandonedPolicyAggregate) UpdatePolicy(cmd command.UpdateTenantCartAbandonedPolicyCommand) error {
	if a.version == -1 {
		return errors.UnpermittedOp.New("tenant policy not created")
	}

	// 変更がない場合はイベントを出さない
	if a.title == cmd.Title &&
		a.cartAbandonedMinutes == cmd.AbandonedMinutes &&
		a.quietTimeFrom.Equal(cmd.QuietTimeFrom) &&
		a.quietTimeTo.Equal(cmd.QuietTimeTo) {
		return nil
	}

	// 直接状態を更新（イベントは後で実装）
	a.title = cmd.Title
	a.cartAbandonedMinutes = cmd.AbandonedMinutes
	a.quietTimeFrom = cmd.QuietTimeFrom
	a.quietTimeTo = cmd.QuietTimeTo
	a.version++

	// TODO: イベントが実装されたら追加
	// ev := event.NewTenantPolicyUpdated(a.tenantID, a.version, title, minutes, quietFrom, quietTo, timezone)
	// a.apply(ev)
	// a.uncommitted = append(a.uncommitted, ev)

	return nil
}
