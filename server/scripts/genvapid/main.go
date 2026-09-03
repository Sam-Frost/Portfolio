//go:build ignore

// genvapid prints a fresh VAPID keypair for Web Push.
//
//	go run ./scripts/genvapid/main.go
//
// Put the output in the server's environment:
//
//	VAPID_PUBLIC_KEY=...
//	VAPID_PRIVATE_KEY=...
//	VAPID_SUBJECT=mailto:you@example.com
//
// The public key is also handed to the browser (GET
// /api/notifications/vapid-public-key) for PushManager.subscribe().
package main

import (
	"fmt"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatalf("generate VAPID keys: %v", err)
	}
	fmt.Printf("VAPID_PUBLIC_KEY=%s\n", pub)
	fmt.Printf("VAPID_PRIVATE_KEY=%s\n", priv)
}
