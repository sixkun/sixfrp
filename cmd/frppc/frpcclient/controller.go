package frpcclient

import (
	"haokun-panel/cmd/frppc/facades"
	"haokun-panel/utils"
)

type clientController struct {
	clients *utils.SyncMap[string, *utils.SyncMap[string, ClientHandler]]
}

func NewClientController() *clientController {
	return &clientController{
		clients: &utils.SyncMap[string, *utils.SyncMap[string, ClientHandler]]{},
	}
}

func (c *clientController) Add(clientID, serverID string, h ClientHandler) {
	m, _ := c.clients.LoadOrStore(clientID, &utils.SyncMap[string, ClientHandler]{})
	if old, loaded := m.LoadAndDelete(serverID); loaded && old != nil {
		old.Stop()
	}
	m.Store(serverID, h)
}

func (c *clientController) Get(clientID, serverID string) ClientHandler {
	v, ok := c.clients.Load(clientID)
	if !ok {
		return nil
	}
	h, ok := v.Load(serverID)
	if !ok {
		return nil
	}
	return h
}

func (c *clientController) Delete(clientID, serverID string) {
	c.Stop(clientID, serverID)
	if v, ok := c.clients.Load(clientID); ok {
		v.Delete(serverID)
	}
}

func (c *clientController) DeleteByClient(clientID string) {
	c.clients.Delete(clientID)
}

func (c *clientController) Run(clientID, serverID string) {
	v, ok := c.clients.Load(clientID)
	if !ok {
		facades.Log().Errorf("controller.Run: client not found, clientID=%s serverID=%s", clientID, serverID)
		return
	}
	h, ok := v.Load(serverID)
	if !ok {
		facades.Log().Errorf("controller.Run: server not found under client, clientID=%s serverID=%s", clientID, serverID)
		return
	}
	go h.Run()
}

func (c *clientController) Stop(clientID, serverID string) {
	v, ok := c.clients.Load(clientID)
	if !ok {
		return
	}
	if h, ok := v.Load(serverID); ok && h != nil {
		h.Stop()
	}
}

func (c *clientController) StopByClient(clientID string) {
	v, ok := c.clients.Load(clientID)
	if !ok {
		return
	}
	v.Range(func(_ string, h ClientHandler) bool {
		if h != nil {
			h.Stop()
		}
		return true
	})
}

func (c *clientController) StopAll() {
	c.clients.Range(func(_ string, v *utils.SyncMap[string, ClientHandler]) bool {
		v.Range(func(_ string, h ClientHandler) bool {
			if h != nil {
				h.Stop()
			}
			return true
		})
		return true
	})
}

func (c *clientController) DeleteAll() {
	c.clients.Range(func(k string, _ *utils.SyncMap[string, ClientHandler]) bool {
		c.DeleteByClient(k)
		return true
	})
	c.clients = &utils.SyncMap[string, *utils.SyncMap[string, ClientHandler]]{}
}

func (c *clientController) List() []string {
	return c.clients.Keys()
}
