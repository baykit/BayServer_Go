package common

type LifecycleListener interface {
	Add(agentId int)
	Remove(agentId int)
}
