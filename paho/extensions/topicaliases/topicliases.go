package topicaliases

import (
	"sync"

	"github.com/blusewang/paho.golang/paho"
)

type TAHandler struct {
	sync.Mutex
	aliasMax uint16
	aliases  []string
}

func NewTAHandler(max uint16) *TAHandler {
	return &TAHandler{
		aliasMax: max + 1,
		aliases:  make([]string, max+1),
	}
}

// GetTopic will return the topic for a given alias number
func (t *TAHandler) GetTopic(a uint16) string {
	if a > t.aliasMax {
		return ""
	}
	t.Lock()
	defer t.Unlock()

	return t.aliases[a]
}

// GetAlias will return the alias for a given topic string
func (t *TAHandler) GetAlias(topic string) uint16 {
	for i, s := range t.aliases {
		if s == topic {
			return uint16(i)
		}
	}

	return 0
}

// SetAlias will request an alias number for a given string
func (t *TAHandler) SetAlias(topic string) uint16 {
	t.Lock()
	defer t.Unlock()

	for i := uint16(1); i <= t.aliasMax; i++ {
		if t.aliases[i] == "" {
			t.aliases[i] = topic
			return i
		}
	}

	return 0
}

// ResetAlias reassigns a given alias number for a new topic
func (t *TAHandler) ResetAlias(topic string, a uint16) {
	t.Lock()
	defer t.Unlock()

	t.aliases[a] = topic
}

// PublishHook is designed to be given to an MQTT client and will be executed
// before a publish is sent allowing it to modify the Properties of the packet.
// In this case it allows the Topic Alias Handler to automatically replace topic
// names with alias numbers
func (t *TAHandler) PublishHook(p *paho.Publish) {
	//p.Topic is always not "" as the default publish checks before calling hooks
	if p.Properties != nil && p.Properties.TopicAlias != nil {
		//topic string is not empty and topic alias is set, reset the alias value.
		t.ResetAlias(p.Topic, *p.Properties.TopicAlias)
		return
	}

	//we already have an alias, set it and unset the topic
	if a := t.GetAlias(p.Topic); a != 0 {
		if p.Properties == nil {
			p.Properties = &paho.PublishProperties{}
		}
		p.Properties.TopicAlias = paho.Uint16(a)
		p.Topic = ""
		return
	}

	//we don't have an alias, try and get one
	if a := t.SetAlias(p.Topic); a != 0 {
		if p.Properties == nil {
			p.Properties = &paho.PublishProperties{}
		}
		p.Properties.TopicAlias = paho.Uint16(a)
		p.Topic = ""
		return
	}
}
