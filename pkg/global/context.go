package global

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ApplicationContext is a struct that holds context and waitgroup for better testability
type ApplicationContext struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

// NewApplicationContext initializes ApplicationContext with once
func NewApplicationContext() *ApplicationContext {
	return sync.OnceValue[*ApplicationContext](newApplicationContext)()
}

// newApplicationContext creates a new ApplicationContext struct with context and waitgroup
func newApplicationContext() *ApplicationContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &ApplicationContext{
		ctx:    ctx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
	}
}

// HandleShutdownSignal listens for system termination signals and cancels the context
func (a *ApplicationContext) HandleShutdownSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, os.Kill, os.Interrupt)
	go func() {
		log.Println("Waiting for termination signal...")
		sig := <-quit
		log.Printf("Received %s signal, initiating shutdown...", sig)
		log.Println("Shutdown signal broadcast")
		a.cancel() // Broadcast cancellation signal to all goroutines
	}()
}

// Context provides the application context and increments the WaitGroup counter
func (a *ApplicationContext) Context() context.Context {
	return a.ctx
}

// Add increments the WaitGroup counter (called when starting a new goroutine)
// !!! Anti-pattern if called inside a goroutine. Only call before WaitForShutdown and Done !!!
func (a *ApplicationContext) Add(delta int) {
	a.wg.Add(delta)
}

// Done decrements the WaitGroup counter, signaling a goroutine has finished
func (a *ApplicationContext) Done() {
	a.wg.Done()
}

// WaitForShutdown waits for all goroutines to finish before allowing main to exit
func (a *ApplicationContext) WaitForShutdown() {
	a.wg.Wait() // Block until all goroutines are done
	log.Println("Gracefully shutting down")
	time.Sleep(time.Millisecond * 3000)
}
