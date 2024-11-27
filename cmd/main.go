package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/libp2p/go-libp2p"
    "github.com/trigg3rX/go-backend/pkg/network"
    "github.com/trigg3rX/triggerx-keeper/execute/handler"
    "github.com/trigg3rX/go-backend/execute/manager"
)

func main() {
    // Setup logging
    log.SetOutput(os.Stdout)
    log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

    // Create context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create libp2p host
    host, err := libp2p.New()
    if err != nil {
        log.Fatalf("Failed to create libp2p host: %v", err)
    }
    defer host.Close()

    // Create network messaging
    messaging := network.NewMessaging(host, "keeper-node-1")

    // Create job handler
    jobHandler := handler.NewJobHandler()

    // Setup message handling
    messaging.InitMessageHandling(func(msg network.Message) {
        job, ok := msg.Content.(*manager.Job)
        if !ok {
            log.Printf("Received invalid job message type")
            return
        }

        if err := jobHandler.HandleJob(job); err != nil {
            log.Printf("Job handling error: %v", err)
        }
    })

    // Save peer info for discovery
    discovery := network.NewDiscovery(ctx, host, "keeper-node-1")
    if err := discovery.SavePeerInfo(); err != nil {
        log.Printf("Failed to save peer info: %v", err)
    }

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    log.Printf("Keeper node started. Listening on %s", host.Addrs())
    <-sigChan
}