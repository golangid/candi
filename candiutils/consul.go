package candiutils

import (
	"encoding/json"
	"time"

	api "github.com/hashicorp/consul/api"
)

// Consul configured for lock acquisition
type Consul struct {
	Client            *api.Client
	Key               string
	SessionID         string
	LockRetryInterval time.Duration
	SessionTTL        time.Duration
}

// ConsulConfig is used to configure creation of client
type ConsulConfig struct {
	ConsulAgentHost   string
	ConsulKey         string
	LockRetryInterval time.Duration
	SessionTTL        time.Duration
}

// NewConsul constructor
func NewConsul(o *ConsulConfig) (*Consul, error) {
	var d Consul
	cfg := api.DefaultConfig()
	cfg.Address = o.ConsulAgentHost
	Client, err := api.NewClient(cfg)
	if err != nil {
		return &d, err
	}

	d.Client = Client
	d.Key = o.ConsulKey
	d.LockRetryInterval = 30 * time.Second
	d.SessionTTL = 5 * time.Minute

	if o.LockRetryInterval != 0 {
		d.LockRetryInterval = o.LockRetryInterval
	}
	if o.SessionTTL != 0 {
		d.SessionTTL = o.SessionTTL
	}

	return &d, nil
}

// RetryLockAcquire attempts to acquire the lock at `LockRetryInterval`
func (d *Consul) RetryLockAcquire(value map[string]string, acquired chan<- struct{}, released chan<- struct{}) {
	ticker := time.NewTicker(d.LockRetryInterval)
	for range ticker.C {
		value["lockAcquisitionTime"] = time.Now().Format(time.RFC3339)
		lock, err := d.acquireLock(value, released)
		if err != nil {
			continue
		}
		if lock {
			ticker.Stop()
			acquired <- struct{}{}
			break
		}
	}
}

// DestroySession method
func (d *Consul) DestroySession() error {
	if d.SessionID == "" {
		return nil
	}
	_, err := d.Client.Session().Destroy(d.SessionID, nil)
	if err != nil {
		return err
	}
	return nil
}

func (d *Consul) createSession() (string, error) {
	return createSession(d.Client, d.Key, d.SessionTTL)
}

// RecreateSession method
func (d *Consul) RecreateSession() error {
	sessionID, err := d.createSession()
	if err != nil {
		return err
	}
	d.SessionID = sessionID
	return nil
}

func (d *Consul) acquireLock(value map[string]string, released chan<- struct{}) (bool, error) {
	if d.SessionID == "" {
		err := d.RecreateSession()
		if err != nil {
			return false, err
		}
	}
	b, _ := json.Marshal(value)
	lockOpts := &api.LockOptions{
		Key:          d.Key,
		Value:        b,
		Session:      d.SessionID,
		LockWaitTime: 1 * time.Millisecond,
		LockTryOnce:  true,
		LockDelay:    1 * time.Millisecond,
		SessionOpts: &api.SessionEntry{
			Behavior:  "release",
			LockDelay: 1 * time.Millisecond,
		},
	}
	lock, err := d.Client.LockOpts(lockOpts)
	if err != nil {
		return false, err
	}
	a, _, err := d.Client.Session().Info(d.SessionID, nil)
	if err == nil && a == nil {
		d.SessionID = ""
		return false, nil
	}
	if err != nil {
		return false, err
	}

	resp, err := lock.Lock(nil)
	if err != nil {
		return false, err
	}
	if resp != nil {
		doneCh := make(chan struct{})
		go func() { d.Client.Session().RenewPeriodic(d.SessionTTL.String(), d.SessionID, nil, doneCh) }()
		go func() {
			<-resp
			close(doneCh)
			released <- struct{}{}
		}()
		return true, nil
	}
	return false, nil
}

func createSession(client *api.Client, consulKey string, ttl time.Duration) (string, error) {
	agentChecks, err := client.Agent().Checks()
	if err != nil {
		return "", err
	}
	checks := []string{}
	checks = append(checks, "serfHealth")
	for _, j := range agentChecks {
		checks = append(checks, j.CheckID)
	}

	sess := &api.SessionEntry{
		Name:      consulKey,
		Checks:    checks,
		LockDelay: 1 * time.Millisecond,
		TTL:       ttl.String(),
	}
	sessionID, _, err := client.Session().Create(sess, nil)
	if err != nil {
		return "", err
	}
	return sessionID, nil
}
