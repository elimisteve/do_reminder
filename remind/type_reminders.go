// Steven Phillips / elimisteve
// 2016.05.30

package remind

import "log"

type Reminders []*Reminder

func (rems Reminders) ByID(id uint64) (*Reminder, error) {
	log.Printf("Searching %d reminders for the one with ID %v\n", len(rems), id)

	for _, r := range rems {
		if r.ID == id {
			return r, nil
		}
	}

	return nil, ErrReminderNotFound
}
