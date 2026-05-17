package model

import (
	"testing"
)

func TestAnalytics_Validate_View(t *testing.T) {
	a := &Analytics{EventType: EventTypeView}
	if err := a.Validate(); err != nil {
		t.Errorf("view event Validate() = %v, want nil", err)
	}
}

func TestAnalytics_Validate_Click(t *testing.T) {
	a := &Analytics{EventType: EventTypeClick}
	if err := a.Validate(); err != nil {
		t.Errorf("click event Validate() = %v, want nil", err)
	}
}

func TestAnalytics_Validate_InvalidEvent(t *testing.T) {
	a := &Analytics{EventType: "hover"}
	if err := a.Validate(); err != ErrInvalidEventType {
		t.Errorf("invalid event Validate() = %v, want ErrInvalidEventType", err)
	}
}

func TestAnalytics_Validate_ValidDeviceTypes(t *testing.T) {
	for _, dt := range []string{DeviceTypeMobile, DeviceTypeTablet, DeviceTypeDesktop} {
		a := &Analytics{EventType: EventTypeView, DeviceType: dt}
		if err := a.Validate(); err != nil {
			t.Errorf("device %q Validate() = %v, want nil", dt, err)
		}
	}
}

func TestAnalytics_Validate_InvalidDeviceType(t *testing.T) {
	a := &Analytics{EventType: EventTypeView, DeviceType: "smartwatch"}
	if err := a.Validate(); err != ErrInvalidDeviceType {
		t.Errorf("invalid device Validate() = %v, want ErrInvalidDeviceType", err)
	}
}

func TestAnalytics_IsView(t *testing.T) {
	a := &Analytics{EventType: EventTypeView}
	if !a.IsView() {
		t.Error("IsView() should return true for view event")
	}
	a.EventType = EventTypeClick
	if a.IsView() {
		t.Error("IsView() should return false for click event")
	}
}

func TestAnalytics_IsClick(t *testing.T) {
	a := &Analytics{EventType: EventTypeClick}
	if !a.IsClick() {
		t.Error("IsClick() should return true for click event")
	}
	a.EventType = EventTypeView
	if a.IsClick() {
		t.Error("IsClick() should return false for view event")
	}
}

func TestAnalyticsSession_Validate_ValidDevice(t *testing.T) {
	as := &AnalyticsSession{DeviceType: DeviceTypeMobile}
	if err := as.Validate(); err != nil {
		t.Errorf("valid device Validate() = %v, want nil", err)
	}
}

func TestAnalyticsSession_Validate_InvalidDevice(t *testing.T) {
	as := &AnalyticsSession{DeviceType: "console"}
	if err := as.Validate(); err != ErrInvalidDeviceType {
		t.Errorf("invalid device Validate() = %v, want ErrInvalidDeviceType", err)
	}
}

func TestAnalyticsSession_Validate_EmptyDevice(t *testing.T) {
	as := &AnalyticsSession{}
	if err := as.Validate(); err != nil {
		t.Errorf("empty device Validate() = %v, want nil", err)
	}
}

func TestAnalyticsSession_IncrementLinkClicks(t *testing.T) {
	as := &AnalyticsSession{LinkClicks: 3}
	as.IncrementLinkClicks()
	if as.LinkClicks != 4 {
		t.Errorf("IncrementLinkClicks() = %d, want 4", as.LinkClicks)
	}
}

func TestAnalyticsSession_UpdateDuration(t *testing.T) {
	as := &AnalyticsSession{}
	as.UpdateDuration(120)
	if as.DurationSeconds != 120 {
		t.Errorf("UpdateDuration(120) = %d, want 120", as.DurationSeconds)
	}
}
