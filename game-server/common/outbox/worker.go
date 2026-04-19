package outbox

/**
* Houses the struct and logic of the workers that process the outbox
* event messages for publishing.
**/

type OutboxWorker struct {
	outboxRepo WorkerOutboxGetter
}

type WorkerOutboxGetter interface {
}

func (w *OutboxWorker) PublishOutboxEvent() error {
}
