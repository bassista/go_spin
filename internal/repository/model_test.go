package repository

import (
	"testing"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestContainer_ApplyDefaults(t *testing.T) {
	c := Container{Name: "test", FriendlyName: "Test", URL: "http://test.local"}
	c.applyDefaults()

	if c.Running == nil {
		t.Error("expected Running to be set")
	}
	if *c.Running != false {
		t.Error("expected Running to default to false")
	}

	if c.Active == nil {
		t.Error("expected Active to be set")
	}
	if *c.Active != false {
		t.Error("expected Active to default to false")
	}
}

func TestContainer_ApplyDefaults_AlreadySet(t *testing.T) {
	c := Container{
		Name:         "test",
		FriendlyName: "Test",
		URL:          "http://test.local",
		Running:      boolPtr(true),
		Active:       boolPtr(true),
	}
	c.applyDefaults()

	if !*c.Running {
		t.Error("expected Running to remain true")
	}
	if !*c.Active {
		t.Error("expected Active to remain true")
	}
}

func TestGroup_ApplyDefaults(t *testing.T) {
	g := Group{Name: "test"}
	g.applyDefaults()

	if g.Container == nil {
		t.Error("expected Container to be initialized")
	}
	if len(g.Container) != 0 {
		t.Error("expected Container to be empty slice")
	}

	if g.Active == nil {
		t.Error("expected Active to be set")
	}
	if *g.Active != false {
		t.Error("expected Active to default to false")
	}
}

func TestSchedule_ApplyDefaults(t *testing.T) {
	s := Schedule{ID: "test", Target: "target", TargetType: "container"}
	s.applyDefaults()

	if s.Timers == nil {
		t.Error("expected Timers to be initialized")
	}
	if len(s.Timers) != 0 {
		t.Error("expected Timers to be empty slice")
	}
}

func TestTimer_ApplyDefaults(t *testing.T) {
	timer := Timer{StartTime: "08:00", StopTime: "18:00"}
	timer.applyDefaults()

	if timer.Active == nil {
		t.Error("expected Active to be set")
	}
	if *timer.Active != false {
		t.Error("expected Active to default to false")
	}

	if timer.Days == nil {
		t.Error("expected Days to be initialized")
	}
	if len(timer.Days) != 0 {
		t.Error("expected Days to be empty slice")
	}
}

func TestDataDocument_ApplyDefaults(t *testing.T) {
	doc := DataDocument{
		Containers: []Container{{Name: "c1", FriendlyName: "C1", URL: "http://c1.local"}},
		Groups:     []Group{{Name: "g1"}},
		Schedules:  []Schedule{{ID: "s1", Target: "c1", TargetType: "container", Timers: []Timer{{StartTime: "08:00", StopTime: "18:00"}}}},
	}

	doc.ApplyDefaults()

	if doc.Containers[0].Running == nil || doc.Containers[0].Active == nil {
		t.Error("expected container defaults to be applied")
	}

	if doc.Groups[0].Active == nil {
		t.Error("expected group defaults to be applied")
	}

	if doc.Schedules[0].Timers[0].Active == nil {
		t.Error("expected timer defaults to be applied")
	}
}

func TestAreDataDocumentsEqual_BothNil(t *testing.T) {
	if !AreDataDocumentsEqual(nil, nil) {
		t.Error("expected nil == nil to be true")
	}
}

func TestAreDataDocumentsEqual_OneNil(t *testing.T) {
	doc := &DataDocument{}
	if AreDataDocumentsEqual(doc, nil) {
		t.Error("expected doc != nil to be false")
	}
	if AreDataDocumentsEqual(nil, doc) {
		t.Error("expected nil != doc to be false")
	}
}

func TestAreDataDocumentsEqual_SameContent(t *testing.T) {
	doc1 := &DataDocument{
		Metadata:   Metadata{LastUpdate: 1000},
		Containers: []Container{{Name: "c1", FriendlyName: "C1", URL: "http://c1.local", Running: boolPtr(false), Active: boolPtr(true)}},
		Order:      []string{"c1"},
	}
	doc2 := &DataDocument{
		Metadata:   Metadata{LastUpdate: 2000}, // Different metadata
		Containers: []Container{{Name: "c1", FriendlyName: "C1", URL: "http://c1.local", Running: boolPtr(false), Active: boolPtr(true)}},
		Order:      []string{"c1"},
	}

	if !AreDataDocumentsEqual(doc1, doc2) {
		t.Error("expected documents with same content (ignoring metadata) to be equal")
	}
}

func TestAreDataDocumentsEqual_DifferentContent(t *testing.T) {
	doc1 := &DataDocument{
		Containers: []Container{{Name: "c1", FriendlyName: "C1", URL: "http://c1.local", Running: boolPtr(false), Active: boolPtr(true)}},
	}
	doc2 := &DataDocument{
		Containers: []Container{{Name: "c2", FriendlyName: "C2", URL: "http://c2.local", Running: boolPtr(false), Active: boolPtr(true)}},
	}

	if AreDataDocumentsEqual(doc1, doc2) {
		t.Error("expected documents with different content to not be equal")
	}
}

func TestAreDataDocumentsEqual_EmptyDocuments(t *testing.T) {
	doc1 := &DataDocument{}
	doc2 := &DataDocument{}

	if !AreDataDocumentsEqual(doc1, doc2) {
		t.Error("expected empty documents to be equal")
	}
}

func TestAreDataDocumentsEqual_DifferentGroups(t *testing.T) {
	doc1 := &DataDocument{
		Groups: []Group{{Name: "g1", Container: []string{"c1"}, Active: boolPtr(true)}},
	}
	doc2 := &DataDocument{
		Groups: []Group{{Name: "g2", Container: []string{"c2"}, Active: boolPtr(false)}},
	}

	if AreDataDocumentsEqual(doc1, doc2) {
		t.Error("expected documents with different groups to not be equal")
	}
}

func TestAreDataDocumentsEqual_DifferentSchedules(t *testing.T) {
	doc1 := &DataDocument{
		Schedules: []Schedule{{ID: "s1", Target: "c1", TargetType: "container"}},
	}
	doc2 := &DataDocument{
		Schedules: []Schedule{{ID: "s2", Target: "c2", TargetType: "group"}},
	}

	if AreDataDocumentsEqual(doc1, doc2) {
		t.Error("expected documents with different schedules to not be equal")
	}
}

func TestAreDataDocumentsEqual_DifferentOrder(t *testing.T) {
	doc1 := &DataDocument{
		Order: []string{"c1", "c2"},
	}
	doc2 := &DataDocument{
		Order: []string{"c2", "c1"},
	}

	if AreDataDocumentsEqual(doc1, doc2) {
		t.Error("expected documents with different order to not be equal")
	}
}

func TestAreDataDocumentsEqual_SameTimers(t *testing.T) {
	doc1 := &DataDocument{
		Schedules: []Schedule{
			{
				ID:         "s1",
				Target:     "c1",
				TargetType: "container",
				Timers: []Timer{
					{StartTime: "08:00", StopTime: "18:00", Days: []int{1, 2, 3}, Active: boolPtr(true)},
				},
			},
		},
	}
	doc2 := &DataDocument{
		Metadata: Metadata{LastUpdate: 9999}, // Different metadata should be ignored
		Schedules: []Schedule{
			{
				ID:         "s1",
				Target:     "c1",
				TargetType: "container",
				Timers: []Timer{
					{StartTime: "08:00", StopTime: "18:00", Days: []int{1, 2, 3}, Active: boolPtr(true)},
				},
			},
		},
	}

	if !AreDataDocumentsEqual(doc1, doc2) {
		t.Error("expected documents with same timers (ignoring metadata) to be equal")
	}
}
