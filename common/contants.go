package common

const (
	EtcdJobPrefix     = "/cron/job/"
	EtcdKillJobPrefix = "/cron/kill/"
	EtcdLockJobPrefix = "/cron/lock/"
)

type EventType uint

const (
	UpdateJob EventType = iota
	DeleteJob
	KillJob
)
