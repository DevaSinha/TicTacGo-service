package match

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

// ── Shared test doubles ────────────────────────────────────────────

type mockPresence struct {
	UserID    string
	SessionID string
	Username  string
	Node      string
}

func (p *mockPresence) GetUserId() string    { return p.UserID }
func (p *mockPresence) GetSessionId() string { return p.SessionID }
func (p *mockPresence) GetNodeId() string    { return p.Node }
func (p *mockPresence) GetHidden() bool      { return false }
func (p *mockPresence) GetPersistence() bool { return true }
func (p *mockPresence) GetUsername() string  { return p.Username }
func (p *mockPresence) GetStatus() string    { return "" }
func (p *mockPresence) GetReason() runtime.PresenceReason {
	return runtime.PresenceReason(0)
}

type mockMatchData struct {
	UserIDVal    string
	SessionIDVal string
	UsernameVal  string
	NodeVal      string
	OpCodeVal    int64
	DataVal      []byte
	ReceiveTime  int64
}

func (m *mockMatchData) GetUserId() string    { return m.UserIDVal }
func (m *mockMatchData) GetSessionId() string { return m.SessionIDVal }
func (m *mockMatchData) GetNodeId() string    { return m.NodeVal }
func (m *mockMatchData) GetHidden() bool      { return false }
func (m *mockMatchData) GetPersistence() bool { return true }
func (m *mockMatchData) GetUsername() string  { return m.UsernameVal }
func (m *mockMatchData) GetStatus() string    { return "" }
func (m *mockMatchData) GetOpCode() int64     { return m.OpCodeVal }
func (m *mockMatchData) GetData() []byte      { return m.DataVal }
func (m *mockMatchData) GetReliable() bool    { return true }
func (m *mockMatchData) GetReceiveTime() int64 { return m.ReceiveTime }
func (m *mockMatchData) GetReason() runtime.PresenceReason {
	return runtime.PresenceReason(0)
}

type mockDispatcher struct {
	BroadcastedOpCode int64
	BroadcastedData   []byte
	BroadcastCount    int
}

func (d *mockDispatcher) BroadcastMessage(opCode int64, data []byte, presences []runtime.Presence, sender runtime.Presence, reliable bool) error {
	d.BroadcastedOpCode = opCode
	d.BroadcastedData = data
	d.BroadcastCount++
	return nil
}

func (d *mockDispatcher) BroadcastMessageDeferred(opCode int64, data []byte, presences []runtime.Presence, sender runtime.Presence, reliable bool) error {
	return nil
}

func (d *mockDispatcher) MatchKick(presences []runtime.Presence) error { return nil }

func (d *mockDispatcher) MatchLabelUpdate(label string) error { return nil }

type mockNakamaModule struct {
	runtime.NakamaModule
	StorageData       map[string]string
	LeaderboardWrites []string
}

func (m *mockNakamaModule) StorageRead(ctx context.Context, reads []*runtime.StorageRead) ([]*api.StorageObject, error) {
	if m.StorageData == nil {
		return nil, nil
	}

	var results []*api.StorageObject
	for _, r := range reads {
		if val, ok := m.StorageData[r.Key]; ok {
			results = append(results, &api.StorageObject{
				Collection: r.Collection,
				Key:        r.Key,
				UserId:     r.UserID,
				Value:      val,
			})
		}
	}
	return results, nil
}

func (m *mockNakamaModule) StorageWrite(ctx context.Context, writes []*runtime.StorageWrite) ([]*api.StorageObjectAck, error) {
	if m.StorageData == nil {
		m.StorageData = make(map[string]string)
	}
	for _, w := range writes {
		m.StorageData[w.Key] = w.Value
	}
	return nil, nil
}

func (m *mockNakamaModule) LeaderboardRecordWrite(ctx context.Context, id string, ownerID string, username string, score int64, subscore int64, metadata map[string]interface{}, overrideOperator *int) (*api.LeaderboardRecord, error) {
	if m.LeaderboardWrites == nil {
		m.LeaderboardWrites = []string{}
	}
	m.LeaderboardWrites = append(m.LeaderboardWrites, ownerID)
	return nil, nil
}

func (m *mockNakamaModule) AccountGetId(ctx context.Context, userID string) (*api.Account, error) {
	return &api.Account{
		User: &api.User{
			Id:       userID,
			Username: "player_" + userID,
		},
	}, nil
}

func (m *mockNakamaModule) LeaderboardCreate(ctx context.Context, id string, authoritative bool, sortOrder string, operator string, resetSchedule string, metadata map[string]interface{}) error {
	return nil
}

func (m *mockNakamaModule) MatchCreate(ctx context.Context, module string, params map[string]interface{}) (string, error) {
	return "match-id-123", nil
}

func (m *mockNakamaModule) LeaderboardRecordsList(ctx context.Context, id string, ownerIDs []string, limit int, cursor string, expiry int64) ([]*api.LeaderboardRecord, []*api.LeaderboardRecord, string, string, error) {
	return []*api.LeaderboardRecord{}, nil, "", "", nil
}

// Compile-time interface checks
var _ runtime.NakamaModule = (*mockNakamaModule)(nil)
var _ runtime.Presence = (*mockPresence)(nil)
var _ runtime.MatchData = (*mockMatchData)(nil)
var _ runtime.MatchDispatcher = (*mockDispatcher)(nil)
var _ runtime.Match = (*Handler)(nil)
var _ = (*sql.DB)(nil)
