// Steven Phillips / elimisteve
// 2016.05.30

package remind

import (
	"log"
	"sync"

	"github.com/boltdb/bolt"
)

type ActiveReminders struct {
	mu        sync.RWMutex
	reminders Reminders
}

func (active *ActiveReminders) All() Reminders {
	active.mu.RLock()
	defer active.mu.RUnlock()

	return active.reminders
}

func (active *ActiveReminders) Cancel(db *bolt.DB, id uint64) error {
	active.mu.Lock()
	defer active.mu.Unlock()

	r, err := active.reminders.ByID(id)
	if err != nil {
		return err
	}

	active.remove(r.ID)

	return r.Cancel(db)
}

func (active *ActiveReminders) Add(rems ...*Reminder) {
	active.mu.Lock()
	defer active.mu.Unlock()

	var notCancelled []*Reminder

	for _, rem := range rems {
		if !rem.Cancelled {
			notCancelled = append(notCancelled, rem)
			continue
		}
		log.Printf("ActiveReminders.Add: Reminder %v cancelled; not adding\n",
			rem.ID)
	}

	active.reminders = append(active.reminders, notCancelled...)
}

func (active *ActiveReminders) remove(id uint64) {
	for i := 0; i < len(active.reminders); i++ {
		if active.reminders[i].ID == id {
			active.reminders = append(active.reminders[:i],
				active.reminders[i+1:]...)
			return
		}
	}
}

func (active *ActiveReminders) Schedule(db *bolt.DB) {
	active.mu.RLock()
	defer active.mu.RUnlock()

	for _, r := range active.reminders {
		go func(r *Reminder) {
			defer func() {
				active.mu.Lock()
				active.remove(r.ID)
				active.mu.Unlock()
				log.Printf("Removed Reminder %v from ActiveReminders\n", r.ID)
			}()

			if err := r.Check(db); err != nil {
				log.Printf("Error scheduling Reminder %v: %v\n", r.ID, err)
				return
			}
			if err := r.RunAndLoop(db); err != nil {
				log.Printf("Error running and looping Reminder %v\n", r.ID, err)
				return
			}
			log.Printf("Reminder %s exited without error\n", r)
		}(r)
	}
}
