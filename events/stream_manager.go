package events

import (
	"sync"

	"github.com/google/uuid"
)

// StreamManager manages client connections and broadcasting.
type StreamManager struct {
	mu                sync.RWMutex
	clientConnections map[string]map[string]*Connection // clientID -> connID -> Connection
}

// NewStreamManager creates a new StreamManager.
func NewStreamManager() *StreamManager {
	return &StreamManager{
		clientConnections: make(map[string]map[string]*Connection),
	}
}

// Connect registers a new client session connection.
func (sm *StreamManager) Connect(session ClientSession) *Connection {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	conn := &Connection{
		ID:      uuid.New().String(),
		Session: session,
		Channel: make(chan []byte, 64),
	}

	conns, ok := sm.clientConnections[session.ClientID]
	if !ok {
		conns = make(map[string]*Connection)
		sm.clientConnections[session.ClientID] = conns
	}
	conns[conn.ID] = conn

	return conn
}

// Disconnect removes a specific connection by its ID.
func (sm *StreamManager) Disconnect(connID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for clientID, conns := range sm.clientConnections {
		if conn, ok := conns[connID]; ok {
			sm.closeConn(conn)
			delete(conns, connID)
			if len(conns) == 0 {
				delete(sm.clientConnections, clientID)
			}
			return
		}
	}
}

// DisconnectAllForClient disconnects all active connections for a given client ID.
func (sm *StreamManager) DisconnectAllForClient(clientID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if conns, ok := sm.clientConnections[clientID]; ok {
		for _, conn := range conns {
			sm.closeConn(conn)
		}
		delete(sm.clientConnections, clientID)
	}
}

// BroadcastToClient sends data to all active connections for a given client ID.
func (sm *StreamManager) BroadcastToClient(clientID string, data []byte) {
	sm.mu.RLock()
	conns, ok := sm.clientConnections[clientID]
	if !ok {
		sm.mu.RUnlock()
		return
	}

	// Copy pointers under RLock to prevent blocking other writes
	connsCopy := make([]*Connection, 0, len(conns))
	for _, conn := range conns {
		connsCopy = append(connsCopy, conn)
	}
	sm.mu.RUnlock()

	for _, conn := range connsCopy {
		sm.sendToConn(conn, data)
	}
}

// sendToConn attempts to send data to a connection's channel in a non-blocking way.
func (sm *StreamManager) sendToConn(conn *Connection, data []byte) {
	conn.mu.RLock()
	if conn.closed {
		conn.mu.RUnlock()
		return
	}
	conn.mu.RUnlock()

	select {
	case conn.Channel <- data:
	default:
		// Drop data if connection buffer is full to prevent slowing down dispatcher
	}
}

// closeConn safely closes a connection's channel.
func (sm *StreamManager) closeConn(conn *Connection) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if !conn.closed {
		close(conn.Channel)
		conn.closed = true
	}
}
