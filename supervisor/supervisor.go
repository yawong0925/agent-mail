// supervisor/supervisor.go
package supervisor

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

// Module defines the contract for all isolated subsystems
type Module interface {
	Name() string
	Start(ctx context.Context, errChan chan<- ModuleError)
	Stop() error
	Status() string
	GetStats() map[string]interface{}
}

// ModuleError wraps an error from a specific module
type ModuleError struct {
	Name string
	Err  error
}

// Supervisor manages module lifecycles and restarts them if they panic
type Supervisor struct {
	modules map[string]Module
	cancels map[string]context.CancelFunc
	mu      sync.RWMutex
}

func New() *Supervisor {
	return &Supervisor{
		modules: make(map[string]Module),
		cancels: make(map[string]context.CancelFunc),
	}
}

func (s *Supervisor) Register(m Module) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modules[m.Name()] = m
}

// Run starts all modules and watches for failures
func (s *Supervisor) Run(ctx context.Context) {
	errChan := make(chan ModuleError, 10)

	// Start all registered modules
	s.mu.RLock()
	for name := range s.modules {
		s.startModule(ctx, name, errChan)
	}
	s.mu.RUnlock()

	// Watcher loop
	for {
		select {
		case <-ctx.Done():
			log.Println("[SUPERVISOR] Shutting down all modules...")
			s.Shutdown()
			return
		case fail := <-errChan:
			log.Printf("[SUPERVISOR] Module '%s' crashed: %v\n", fail.Name, fail.Err)
			log.Printf("[SUPERVISOR] Restarting '%s' in 5 seconds...\n", fail.Name)
			time.Sleep(5 * time.Second)
			
			// Only restart if context isn't cancelled during the sleep
			if ctx.Err() == nil {
				s.startModule(ctx, fail.Name, errChan)
			}
		}
	}
}

func (s *Supervisor) startModule(parentCtx context.Context, name string, errChan chan<- ModuleError) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mod := s.modules[name]
	modCtx, cancel := context.WithCancel(parentCtx)
	s.cancels[name] = cancel

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- ModuleError{
					Name: name,
					Err:  fmt.Errorf("panic: %v\n%s", r, debug.Stack()),
				}
			}
		}()
		
		log.Printf("[SUPERVISOR] Starting module: %s\n", name)
		mod.Start(modCtx, errChan)
	}()
}

// supervisor/supervisor.go (replace the Shutdown function)

func (s *Supervisor) Shutdown() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	log.Println("[SUPERVISOR] Reached shutdown checkpoint: Initiating cascade shutdown of all modules...")
	
	for name, cancel := range s.cancels {
		log.Printf("[SUPERVISOR] Reached shutdown checkpoint: Sending context cancel signal to '%s'\n", name)
		cancel() // Signal the module's context to stop
		
		log.Printf("[SUPERVISOR] Reached shutdown checkpoint: Calling Stop() on '%s'\n", name)
		if err := s.modules[name].Stop(); err != nil {
			log.Printf("[SUPERVISOR] Error stopping '%s': %v\n", name, err)
		}
		
		log.Printf("[SUPERVISOR] '%s' reached complete shutdown.\n", name)
	}
	
	log.Println("[SUPERVISOR] All modules reached complete shutdown.")
}