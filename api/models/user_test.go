package models

import "testing"

func TestNotificationPrefs_EffectiveDefaultsToAllOn(t *testing.T) {
	// All four fields nil — the "user never touched their settings" state.
	got := NotificationPrefs{}.Effective()
	if !got.InAppNewVuln || !got.InAppNewComment || !got.PushNewVuln || !got.PushNewComment {
		t.Fatalf("default prefs should be all-on, got %+v", got)
	}
}

func TestNotificationPrefs_EffectiveRespectsExplicitOff(t *testing.T) {
	off := false
	on := true
	got := NotificationPrefs{
		InAppNewVuln:    &off,
		InAppNewComment: &on,
		PushNewVuln:     nil,    // unset -> default on
		PushNewComment:  &off,
	}.Effective()
	if got.InAppNewVuln {
		t.Fatalf("explicit false should win for InAppNewVuln")
	}
	if !got.InAppNewComment {
		t.Fatalf("explicit true should win for InAppNewComment")
	}
	if !got.PushNewVuln {
		t.Fatalf("nil should default to true for PushNewVuln")
	}
	if got.PushNewComment {
		t.Fatalf("explicit false should win for PushNewComment")
	}
}
