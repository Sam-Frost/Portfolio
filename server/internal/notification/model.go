package notification

import "time"

// PushSubscription is one browser's Web Push registration, as produced by
// PushManager.subscribe() on the client. Endpoint is the push service URL
// and the natural unique key; P256dh and Auth are the client's payload
// encryption keys.
type PushSubscription struct {
	ID        string    `json:"id"`
	Endpoint  string    `json:"endpoint"`
	P256dh    string    `json:"p256dh"`
	Auth      string    `json:"auth"`
	UserAgent *string   `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
}

// SubscribeInput is the client-supplied subscription payload.
type SubscribeInput struct {
	Endpoint  string  `json:"endpoint"`
	P256dh    string  `json:"p256dh"`
	Auth      string  `json:"auth"`
	UserAgent *string `json:"userAgent"`
}

// Message is one notification to deliver across every enabled channel
// (email + Web Push). URL is the in-app path a push click should open
// (e.g. "/todos"); Tag coalesces same-tag notifications on the device.
type Message struct {
	Title string
	Body  string
	URL   string
	Tag   string
}

// Log kinds for the dedup ledger (notification_log).
const KindMorningDigest = "morning_digest"
