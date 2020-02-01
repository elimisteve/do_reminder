// Steven Phillips / elimisteve
// 2016.05.30

package remind

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/boltdb/bolt"
)

type ActiveReminders struct {
	mu        sync.RWMutex
	reminders Reminders
}

func (active *ActiveReminders) Cancel(db *bolt.DB, ids []uint64) error {
	active.mu.Lock()
	defer active.mu.Unlock()

	errs := []string{}

	for _, id := range ids {
		r, err := active.reminders.ByID(id)
		if err != nil {
			errs = append(errs, fmt.Sprintf(
				"Error getting reminder #%d: %v", id, err.Error()))
			continue
		}

		active.remove(r.ID)

		cancErr := r.Cancel(db)
		if cancErr != err {
			errs = append(errs, err.Error())
			continue
		}
	}

	if len(errs) == 1 {
		return errors.New(errs[0])
	} else if len(errs) != 0 {
		errStr := errs[0]
		for i := 1; i < len(errs); i++ {
			errStr += "; " + errs[i]
		}
		return fmt.Errorf("Got %d errors: %v", len(errs), errStr)
	}

	return nil
}

func (active *ActiveReminders) add(rems ...*Reminder) {
	active.reminders = append(active.reminders, rems...)
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

func (active *ActiveReminders) Schedule(db *bolt.DB, rems Reminders) {
	rems = rems.NotCancelled()
	active.add(rems...)
	for _, r := range rems {
		go func(r *Reminder) {
			if err := active.manage(db, r); err != nil {
				log.Printf("Scheduled Reminder %v exited with error: %v\n",
					r.ID, err)
			}
		}(r)
	}
}

func (active *ActiveReminders) manage(db *bolt.DB, r *Reminder) error {
	defer func() {
		active.mu.Lock()
		active.remove(r.ID)
		active.mu.Unlock()
		log.Printf("Removed Reminder %v from ActiveReminders\n", r.ID)
	}()

	if err := r.Check(db); err != nil {
		return fmt.Errorf("Error scheduling Reminder %v: %v\n", r.ID, err)
	}
	if err := r.RunAndLoop(db); err != nil {
		return fmt.Errorf("Error running and looping Reminder %v: %v\n", r.ID,
			err)
	}

	log.Printf("Reminder %s exited without error\n", r)
	return nil
}

func (active *ActiveReminders) ScheduleNew(db *bolt.DB, r *Reminder) error {
	active.add(r)

	if err := r.Check(db); err != nil {
		active.mu.Lock()
		active.remove(r.ID)
		active.mu.Unlock()
		return fmt.Errorf("Error checking Reminder %v: %v\n", r.ID, err)
	}

	go func() {
		if err := r.RunAndLoop(db); err != nil {
			log.Printf("Error running and looping Reminder %v: %v\n", r.ID, err)
		}
		active.mu.Lock()
		active.remove(r.ID)
		active.mu.Unlock()
	}()

	return nil
}
