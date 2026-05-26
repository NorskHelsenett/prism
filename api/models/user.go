package models

// UserSettings is persisted as a JSON blob in user_data.settings. Unknown keys
// are tolerated on read so older clients can deserialize newer payloads.
type UserSettings struct {
	SwimlaneUsers     []string          `json:"swimlaneUsers" gorm:"swimlaneUsers"`
	NotificationPrefs NotificationPrefs `json:"notificationPrefs"`
}

// NotificationPrefs is the per-user opt-out matrix for the two channels and
// two event kinds. All four flags default to true (notifications enabled);
// the dispatcher only suppresses delivery when a flag has been explicitly set
// to false. Unset fields decode as false in Go but are normalised to true by
// EffectiveNotificationPrefs at read time.
type NotificationPrefs struct {
	InAppNewVuln    *bool `json:"inAppNewVuln,omitempty"`
	InAppNewComment *bool `json:"inAppNewComment,omitempty"`
	PushNewVuln     *bool `json:"pushNewVuln,omitempty"`
	PushNewComment  *bool `json:"pushNewComment,omitempty"`
}

// EffectiveNotificationPrefs replaces nil pointers with `true`, giving the
// dispatcher a flag-set with no ambiguity between "unset" and "off". A user
// who has never touched their settings gets every notification.
func (p NotificationPrefs) Effective() ResolvedNotificationPrefs {
	return ResolvedNotificationPrefs{
		InAppNewVuln:    boolOrTrue(p.InAppNewVuln),
		InAppNewComment: boolOrTrue(p.InAppNewComment),
		PushNewVuln:     boolOrTrue(p.PushNewVuln),
		PushNewComment:  boolOrTrue(p.PushNewComment),
	}
}

type ResolvedNotificationPrefs struct {
	InAppNewVuln    bool
	InAppNewComment bool
	PushNewVuln     bool
	PushNewComment  bool
}

func boolOrTrue(b *bool) bool {
	if b == nil {
		return true
	}
	return *b
}
