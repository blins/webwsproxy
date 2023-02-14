package main

import "sync"

type Channel struct {
	Name     string
	sockets  []*WS
	msg_chan chan []byte
	mux      sync.RWMutex
}

// вернуть последний значимый индекс в массиве
func (c *Channel) lastindex() int {
	l := len(c.sockets)
	for i := l - 1; i > -1; i-- {
		if c.sockets[i] != nil && c.sockets[i].Alive() {
			return i
		}
	}
	return -1
}

func (c *Channel) loop() {
	for {
		select {
		case msg, ok := <-c.msg_chan:
			if !ok {
				return
			}
			c.mux.Lock()
			for i := 0; i < len(c.sockets); i++ {
				// garbage
				if c.sockets[i] == nil || !c.sockets[i].Alive() {
					l := c.lastindex()
					if i >= l {
						break
					}
					if c.sockets[i] != nil {
						c.sockets[i] = nil
					}
					c.sockets[i] = c.sockets[l]
					c.sockets[l] = nil
				}
				c.sockets[i].Send(msg)
			}
			c.mux.Unlock()
		}
	}
}

func (c *Channel) AddWS(ws *WS) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.sockets == nil {
		c.sockets = make([]*WS, 0)
	}
	l := len(c.sockets)
	li := c.lastindex()
	if li < l-1 {
		c.sockets[li+1] = ws
	} else {
		c.sockets = append(c.sockets, ws)
	}
}

func (c *Channel) Send(msg []byte) {
	c.msg_chan <- msg
}

func (c *Channel) Close() error {
	close(c.msg_chan)
	for _, w := range c.sockets {
		if w != nil {
			w.Close()
		}
	}
	return nil
}

func (c *Channel) Run() {
	if c.msg_chan == nil {
		c.msg_chan = make(chan []byte, 10)
	}
	go c.loop()
}

type Channels struct {
	channel map[string]*Channel
	mux     sync.RWMutex
}

func (c *Channels) Add(name string, ws *WS) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.channel == nil {
		c.channel = make(map[string]*Channel)
	}
	if ch, ok := c.channel[name]; ok {
		ch.AddWS(ws)
	} else {
		ch = &Channel{
			Name: name,
		}
		ch.AddWS(ws)
		ch.Run()
		c.channel[name] = ch
	}
}

func (c *Channels) Close() error {
	for _, v := range c.channel {
		v.Close()
	}
	return nil
}

func (c *Channels) Send(name string, msg []byte) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if c.channel == nil {
		return
	}
	if ch, ok := c.channel[name]; ok {
		ch.Send(msg)
	}
}
