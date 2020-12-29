package candiutils

import (
	"encoding/json"
	"net/url"
	"time"

	api "github.com/hashicorp/consul/api"
	"pkg.agungdwiprasetyo.com/candi/logger"
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
func NewConsul(opt *ConsulConfig) (*Consul, error) {
	var c Consul
	cfg := api.DefaultConfig()
	cfg.Address = opt.ConsulAgentHost
	Client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c.Client = Client
	c.Key = opt.ConsulKey
	c.LockRetryInterval = 30 * time.Second
	c.SessionTTL = 5 * time.Minute

	if opt.LockRetryInterval != 0 {
		c.LockRetryInterval = opt.LockRetryInterval
	}
	if opt.SessionTTL != 0 {
		c.SessionTTL = opt.SessionTTL
	}

	return &c, nil
}

// RetryLockAcquire attempts to acquire the lock at `LockRetryInterval`
func (c *Consul) RetryLockAcquire(value map[string]string, acquired chan<- struct{}, released chan<- struct{}) {
	ticker := time.NewTicker(c.LockRetryInterval)
	for range ticker.C {
		value["lockAcquisitionTime"] = time.Now().Format(time.RFC3339)
		lock, err := c.acquireLock(value, released)
		if err != nil {
			logger.LogYellow("Cannot connect to consul, " + err.Error())
			switch err.(type) {
			case *url.Error:
				goto ACQUIRED
			default:
				continue
			}
		}
		if lock {
			goto ACQUIRED
		}
	}

ACQUIRED:
	ticker.Stop()
	acquired <- struct{}{}
}

// DestroySession method
func (c *Consul) DestroySession() error {
	if c.SessionID == "" {
		return nil
	}
	_, err := c.Client.Session().Destroy(c.SessionID, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Consul) createSession() (string, error) {
	return createSession(c.Client, c.Key, c.SessionTTL)
}

// RecreateSession method
func (c *Consul) RecreateSession() error {
	sessionID, err := c.createSession()
	if err != nil {
		return err
	}
	c.SessionID = sessionID
	return nil
}

func (c *Consul) acquireLock(value map[string]string, released chan<- struct{}) (bool, error) {
	if c.SessionID == "" {
		err := c.RecreateSession()
		if err != nil {
			return false, err
		}
	}
	b, _ := json.Marshal(value)
	lockOpts := &api.LockOptions{
		Key:          c.Key,
		Value:        b,
		Session:      c.SessionID,
		LockWaitTime: 1 * time.Millisecond,
		LockTryOnce:  true,
		LockDelay:    1 * time.Millisecond,
		SessionOpts: &api.SessionEntry{
			Behavior:  "release",
			LockDelay: 1 * time.Millisecond,
		},
	}
	lock, err := c.Client.LockOpts(lockOpts)
	if err != nil {
		return false, err
	}
	a, _, err := c.Client.Session().Info(c.SessionID, nil)
	if err == nil && a == nil {
		c.SessionID = ""
		return false, nil
	}
	if err != nil {
		return false, err
	}

	unlock, err := lock.Lock(nil)
	if err != nil {
		return false, err
	}
	if unlock != nil {
		doneCh := make(chan struct{})
		go func() { c.Client.Session().RenewPeriodic(c.SessionTTL.String(), c.SessionID, nil, doneCh) }()
		go func() {
			<-unlock
			close(doneCh)
			time.Sleep(c.LockRetryInterval)
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
