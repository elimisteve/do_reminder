// Steven Phillips / elimisteve
// 2016.05.28

package remind

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/elimisteve/do_reminder/twilhelp"
)

var (
	boltBucket = []byte("reminder")

	ErrReminderNotFound = errors.New("Reminder not found")
)

type Reminder struct {
	ID          uint64
	Recipient   string
	Description string
	NextRun     time.Time
	Period      time.Duration // Period == 0 means should only run once
	PlusMinus   time.Duration

	Raw     string
	Created time.Time

	Cancelled bool
	cancel    chan struct{}
}

type Reminders []*Reminder

func (rems Reminders) Schedule(db *bolt.DB) []error {
	var errs []error
	for _, r := range rems {
		if r.Cancelled {
			continue
		}
		if err := r.Schedule(db); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func GetAllReminders(db *bolt.DB) (Reminders, error) {
	var allRems Reminders

	e := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(boltBucket)
		if err != nil {
			return err
		}

		return b.ForEach(func(k, v []byte) error {
			var rem Reminder

			if err := json.Unmarshal(v, &rem); err != nil {
				return err
			}

			rem.ID = binary.BigEndian.Uint64(k)

			allRems = append(allRems, &rem)

			return nil
		})
	})

	return allRems, e
}

// Schedule reminds r.Recipient to do r.Description starting at
// r.NextRun, then every r.Period +/- r.PlusMinus after that.
func (r *Reminder) Schedule(db *bolt.DB) error {
	if r == nil {
		errors.New("Cannot schedule nil *Reminder!")
	}

	if r.Period < 0 {
		return fmt.Errorf("Reminder cannot have negative period (%v)", r.Period)
	}

	log.Printf("Valid Reminder %v scheduled\n", r.ID)

	if r.NextRun.Before(Now()) {
		if r.Period == 0 {
			log.Printf("Reminder %#v's next run already passed, should have"+
				" only run once; returning nil\n", r)
			r.Cancelled = true
			return r.Update(db)
		}
		changed, err := r.FutureizeNextRun()
		if err != nil {
			return err
		}
		if changed {
			if err := r.Update(db); err != nil {
				return err
			}
		}
	}

	go func() {
		if err := r.RunAndLoop(db); err != nil {
			log.Printf("Error running looping reminder %v: %v\n", r.ID, err)
			return
		}
		log.Printf("Reminder %v stopped looping (no error)\n", r.ID)
	}()

	return nil
}

func (r *Reminder) RunAndLoop(db *bolt.DB) error {
	r.NextRun = r.NextRun.Add(RandDuration(r.PlusMinus))
	if r.PlusMinus != 0 {
		if err := r.Update(db); err != nil {
			return fmt.Errorf("Error updating reminder: %v", err)
		}
	}

	// Sleep till the next run is here
	dur := max(r.NextRun.Sub(Now()), 0)

	select {
	case <-r.cancel:
		log.Printf("Reminder %#v cancelled; returning\n", r)
		return nil
	case <-time.After(dur):
		// Keep going
	}

	for {
		log.Printf("Texting `%s` to remind him/her to `%s` starting now then every ~%s after that\n",
			r.Recipient, r.Description, r.Period)

		err := twilhelp.SendSMS(r.Recipient, r.Description)
		if err != nil {
			log.Printf("Error sending SMS `%v` to `%v`: %v\n", r.Description,
				r.Recipient, err)

			// TODO: Return?
			time.Sleep(1 * time.Second)
		}

		if r.Period == 0 {
			if err != nil {
				log.Printf("PROBLEM: Reminder %#v should only send once, but "+
					"failed to send; erroring out, not trying again\n", r)
				return err
			}
			log.Printf("Reminder %v successfully ran once; exiting\n", r.ID)
			r.Cancelled = true
			return r.Update(db)
		}

		// TODO: Prevent drift. Right now there's nothing stopping
		// the time at which a reminder runs from drifting 60 mins
		// every single time!
		sleep := r.Period + RandDuration(r.PlusMinus)
		log.Printf("Text to %s, `%s`, sending again in %s (period: %s)\n",
			r.Recipient, r.Description, sleep, r.Period)

		r.NextRun = Now().Add(sleep)
		if err := r.Update(db); err != nil {
			return err
		}

		select {
		case <-r.cancel:
			log.Printf("Reminder %#v cancelled; returning\n", r)
			return nil
		case <-time.After(max(sleep, -sleep)):
			// Keep going
		}
	}
}

// Set r.NextRun to be in the future
func (r *Reminder) FutureizeNextRun() (changed bool, err error) {
	if r.Period == 0 {
		return false, errors.New("Cannot futurize reminder with a period of 0")
	}
	now := Now()
	if r.NextRun.After(now) {
		return false, nil
	}
	future := r.Period + now.Sub(r.NextRun)/r.Period

	r.NextRun = r.NextRun.Add(future)

	return true, nil
}

func max(n, m time.Duration) time.Duration {
	if n > m {
		return n
	}
	return m
}

func (r *Reminder) Save(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(boltBucket)
		if err != nil {
			return err
		}

		id, _ := b.NextSequence()
		r.ID = id

		rBytes, err := json.Marshal(r)
		if err != nil {
			return err
		}

		return b.Put(r.IDBytes(), rBytes)
	})
}

func (r *Reminder) Update(db *bolt.DB) error {
	if r == nil {
		return errors.New("Cannot update nil *Reminder")
	}

	if db == nil {
		return fmt.Errorf("Error updating Reminder %v; db is nil", r.ID)
	}

	log.Printf("Updating reminder %v\n", r.ID)

	return db.Update(func(tx *bolt.Tx) error {
		rBytes, err := json.Marshal(r)
		if err != nil {
			return err
		}

		b := tx.Bucket(boltBucket)

		return b.Put(r.IDBytes(), rBytes)
	})
}

func (r *Reminder) Cancel(db *bolt.DB) error {
	r.cancel <- struct{}{}
	r.Cancelled = true

	err := r.Update(db)
	if err != nil {
		return fmt.Errorf(
			"Cancelled currently-running Reminder, but failed to save: %v", err)
	}
	return nil
}

func (r *Reminder) IDBytes() []byte {
	return itob(r.ID)
}

// From https://github.com/boltdb/bolt#autoincrementing-integer-for-the-bucket
//
// itob returns an 8-byte big endian representation of id
func itob(id uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, id)
	return b
}
