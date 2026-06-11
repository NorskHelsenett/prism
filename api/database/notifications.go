package database

import (
	"errors"
	"time"

	"prism/models"

	"gorm.io/gorm"
)

// CreateNotification inserts one row into notifications and returns the
// stored row in wire form (id and createdAt populated) so the dispatcher can
// stream it to connected clients. This is a single INSERT — there is no
// read-modify-write — so two dispatchers can race to write notifications for
// the same user without losing data, which was the failure mode of the old
// UserData.Notifications JSON column.
func CreateNotification(recipientEmail string, n models.Notification) (models.Notification, error) {
	if recipientEmail == "" {
		// Defensive — the dispatcher already filters empties, but cheap to
		// short-circuit rather than insert a row no one can read.
		return models.Notification{}, nil
	}
	row := Notification{
		RecipientEmail:  recipientEmail,
		Kind:            n.Kind,
		Who:             n.Who,
		What:            n.What,
		Where:           n.Where,
		VulnerabilityID: n.VulnerabilityID,
		IsRead:          false,
		CreatedAt:       time.Now().UTC(),
	}
	if err := db.Create(&row).Error; err != nil {
		return models.Notification{}, err
	}
	return row.Wire(), nil
}

// GetNotifications returns the most recent notifications for one user. The
// dropdown only renders a window of recent rows; older history stays in the
// DB and is accessible via the same endpoint with a higher limit if needed.
func GetNotifications(recipientEmail string, limit int) ([]models.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []Notification
	err := db.Where("recipient_email = ?", recipientEmail).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]models.Notification, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Wire())
	}
	return out, nil
}

// MarkNotificationAsRead flips IsRead for one row, scoped by recipient so one
// user can never mark another user's row.
func MarkNotificationAsRead(recipientEmail string, id uint) error {
	res := db.Model(&Notification{}).
		Where("id = ? AND recipient_email = ?", id, recipientEmail).
		Update("is_read", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAllNotificationsRead flips IsRead on every unread row for the user.
// Replaces the old "delete everything" flow that the dropdown used to call
// for clearing the badge.
func MarkAllNotificationsRead(recipientEmail string) error {
	return db.Model(&Notification{}).
		Where("recipient_email = ? AND is_read = ?", recipientEmail, false).
		Update("is_read", true).Error
}

// DeleteNotifications removes every row for one user. Kept for compatibility
// with the existing DELETE /api/notification endpoint, but the dropdown now
// uses MarkAllNotificationsRead so history survives "clear all".
func DeleteNotifications(recipientEmail string) error {
	if recipientEmail == "" {
		return errors.New("recipient email required")
	}
	return db.Where("recipient_email = ?", recipientEmail).Delete(&Notification{}).Error
}

// Wire converts the DB row into the API DTO consumed by the frontend
// dropdown. Keeps the wire shape stable while letting the DB schema evolve.
func (n *Notification) Wire() models.Notification {
	return models.Notification{
		ID:              n.ID,
		Who:             n.Who,
		What:            n.What,
		IsRead:          n.IsRead,
		Where:           n.Where,
		When:            n.CreatedAt,
		Kind:            n.Kind,
		VulnerabilityID: n.VulnerabilityID,
	}
}
