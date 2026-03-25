package linter

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
)

type summary struct {
	mu           sync.Mutex
	numError     atomic.Int32
	numWarning   atomic.Int32
	numJob       atomic.Int32
	numProcessed atomic.Int32
	glitches     []*glitch.Glitch
}

func newSummary() *summary {
	return &summary{
		numError:     atomic.Int32{},
		numWarning:   atomic.Int32{},
		numJob:       atomic.Int32{},
		numProcessed: atomic.Int32{},
	}
}

func (s *summary) addGlitch(g *glitch.Glitch) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.glitches = append(s.glitches, g)
}

func (s *summary) markdown(title string, limit int) string {
	markdown := fmt.Sprintf("# %s\n", title)

	if len(s.glitches) > 0 {
		glitchesMd := glitch.ListToMarkdown(s.glitches, limit)
		markdown += "Found non-compliant files:\n\n"
		markdown += glitchesMd
	} else {
		markdown += "No errors or warning were found\n\n"
	}

	return markdown
}
