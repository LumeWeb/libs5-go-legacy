package protocol

import (
	"github.com/golang/mock/gomock"
	"os"
	"testing"
)

// Common resources
var (
	mockCtrl *gomock.Controller
	node     *interfaces.MockNode
	peer     *net.MockPeer
	services *interfaces.MockServices
	p2p      *interfaces.MockP2PService
)

// Setup function
func setup(t *testing.T) {
	mockCtrl = gomock.NewController(t)
	node = interfaces.NewMockNode(mockCtrl)
	peer = net.NewMockPeer(mockCtrl)
	services = interfaces.NewMockServices(mockCtrl)
	p2p = interfaces.NewMockP2PService(mockCtrl)
}

// Teardown function
func teardown() {
	mockCtrl.Finish()
	// Other cleanup tasks
}

// TestMain function for setup and teardown
func TestMain(m *testing.M) {
	code := m.Run()
	teardown()
	os.Exit(code)
}
