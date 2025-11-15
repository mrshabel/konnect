package worker

// queue definitions
const (
	CriticalQueue = "critical"
	EmailQueue    = "email"
	SMSQueue      = "sms"
	InAppQueue    = "inapp"
	DefaultQueue  = "default"
	LowQueue      = "low"
)

// worker queues with priority assignment
var Queues = map[string]int{
	CriticalQueue: 6,
	EmailQueue:    3,
	SMSQueue:      3,
	InAppQueue:    5,
	DefaultQueue:  3,
	LowQueue:      1,
}
