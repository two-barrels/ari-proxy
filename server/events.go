package server

import "github.com/two-barrels/ari/v6"

func (s *Server) dialogsForEvent(e ari.Event) (ret []string) {
	seen := make(map[string]bool)
	for _, k := range e.Keys() {
		if k == nil {
			s.Log.Warn("received nil key for event", "event", e)
			continue
		}
		for _, dialog := range s.Dialog.List(k.Kind, k.ID) {
			if !seen[dialog] {
				ret = append(ret, dialog)
				seen[dialog] = true
			}
		}
	}
	return
}
