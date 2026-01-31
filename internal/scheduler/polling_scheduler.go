package scheduler

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/bassista/go_spin/internal/runtime"
)

type DayFlags struct {
	StartedDayKey string
	StoppedDayKey string
}

// PollingScheduler evaluates schedules on a fixed interval and performs at most
// one start and one stop action per container per day (in the configured timezone).
//
// Semantics:
// - If StartedDayKey == today, start is never attempted again today, regardless of running state.
// - If StoppedDayKey == today, stop is never attempted again today.
// - Stop evaluation is only performed after a start evaluation has happened that day.
//
// NOTE: Flags are in-memory only.
type PollingScheduler struct {
	store   cache.ReadOnlyStore
	runtime runtime.ContainerRuntime
	poll    time.Duration
	loc     *time.Location
	logger  *log.Logger

	mu    sync.Mutex
	flags map[string]DayFlags
}

func NewPollingScheduler(store cache.ReadOnlyStore, rt runtime.ContainerRuntime, poll time.Duration, loc *time.Location) *PollingScheduler {
	if loc == nil {
		loc = time.Local
	}

	return &PollingScheduler{
		store:   store,
		runtime: rt,
		poll:    poll,
		loc:     loc,
		logger:  log.New(os.Stdout, "[sched] ", log.LstdFlags),
		flags:   map[string]DayFlags{},
	}
}

func (s *PollingScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.poll)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				s.logger.Println("scheduler stopped")
				return
			case <-ticker.C:
				s.tick(ctx)
			}
		}
	}()
}

func (s *PollingScheduler) tick(ctx context.Context) {
	doc, err := s.store.Snapshot()
	if err != nil {
		s.logger.Printf("snapshot error: %v", err)
		return
	}

	now := time.Now().In(s.loc)
	todayKey := dayKey(now)

	// Build lookup maps for efficient access during schedule evaluation.
	containersByName := map[string]repository.Container{}
	for _, c := range doc.Containers {
		if c.Name == "" {
			continue
		}
		containersByName[c.Name] = c
	}

	groupsByName := map[string]repository.Group{}
	for _, g := range doc.Groups {
		if g.Name == "" {
			continue
		}
		groupsByName[g.Name] = g
	}

	// Initialize desiredRunning map: by default, no container should be running.
	// This will be set to true if any active schedule/timer indicates it should be running now.
	desiredRunning := map[string]bool{}
	for name := range containersByName {
		desiredRunning[name] = false
	}

	// Evaluate all schedules to determine which containers should be running based on active timers.
	for _, sched := range doc.Schedules {
		// Expand the schedule target into a list of container names (handles both "container" and "group" target types).
		containerNames := expandScheduleTargets(sched, containersByName, groupsByName)
		if len(containerNames) == 0 {
			continue
		}

		for _, timer := range sched.Timers {
			if timer.Active != nil && !*timer.Active {
				continue
			}
			// Check if this timer is currently active (within its start/stop window, considering days and cross-midnight).
			if !isTimerActiveNow(timer, now) {
				continue
			}

			// For each container targeted by this schedule, mark it as desired running if the container itself is active.
			for _, containerName := range containerNames {
				c, ok := containersByName[containerName]
				if !ok {
					continue
				}
				// Respect the container's own active flag.
				if c.Active != nil && !*c.Active {
					continue
				}
				desiredRunning[containerName] = true
			}
		}
	}

	// For each container, decide whether to start or stop based on desired state and day-key flags.
	for containerName := range containersByName {
		flags := s.getFlags(containerName)
		shouldRun := desiredRunning[containerName]
		// If we already attempted to start this container today, skip to avoid repeated attempts.
		// This enforces "at most one start per day" even if the container stops later.
		if shouldRun {
			if flags.StartedDayKey == todayKey {
				continue
			}
			// Check current runtime state.
			running, err := s.runtime.IsRunning(ctx, containerName)
			if err != nil {
				s.logger.Printf("IsRunning(%s) error: %v", containerName, err)
				continue
			}
			if !running {
				if err := s.runtime.Start(ctx, containerName); err != nil {
					s.logger.Printf("Start(%s) error: %v", containerName, err)
					continue
				}
				s.logger.Printf("started %s", containerName)
			}
			// Mark that a start attempt was made today (even if it was already running).
			flags.StartedDayKey = todayKey
			s.setFlags(containerName, flags)
			continue
		}

		// Container should not be running now.
		// Stop evaluation only happens if a start evaluation occurred today (to avoid premature stops).
		if flags.StartedDayKey != todayKey {
			// Stop action is only evaluated after a start evaluation has happened today.
			continue
		}
		// If we already attempted to stop this container today, skip.
		if flags.StoppedDayKey == todayKey {
			continue
		}

		running, err := s.runtime.IsRunning(ctx, containerName)
		if err != nil {
			s.logger.Printf("IsRunning(%s) error: %v", containerName, err)
			continue
		}
		if running {
			if err := s.runtime.Stop(ctx, containerName); err != nil {
				s.logger.Printf("Stop(%s) error: %v", containerName, err)
				continue
			}
			s.logger.Printf("stopped %s", containerName)
		}
		// Mark that a stop attempt was made today (even if it was already stopped).
		flags.StoppedDayKey = todayKey
		s.setFlags(containerName, flags)
	}
}

func (s *PollingScheduler) getFlags(containerName string) DayFlags {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flags[containerName]
}

func (s *PollingScheduler) setFlags(containerName string, flags DayFlags) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flags[containerName] = flags
}

func dayKey(t time.Time) string {
	return t.Format("2006-01-02")
}

func expandScheduleTargets(
	sched repository.Schedule,
	containersByName map[string]repository.Container,
	groupsByName map[string]repository.Group,
) []string {
	if sched.Target == "" {
		return nil
	}

	switch sched.TargetType {
	case "container":
		if _, ok := containersByName[sched.Target]; !ok {
			return nil
		}
		return []string{sched.Target}
	case "group":
		g, ok := groupsByName[sched.Target]
		if !ok {
			return nil
		}
		if g.Active != nil && !*g.Active {
			return nil
		}
		out := make([]string, 0, len(g.Container))
		for _, name := range g.Container {
			if name == "" {
				continue
			}
			out = append(out, name)
		}
		return out
	default:
		return nil
	}
}

func isTimerActiveNow(timer repository.Timer, now time.Time) bool {
	startClock, err := time.Parse("15:04", timer.StartTime)
	if err != nil {
		return false
	}
	stopClock, err := time.Parse("15:04", timer.StopTime)
	if err != nil {
		return false
	}

	// Check windows anchored to today and yesterday (handles cross-midnight).
	for _, dayOffset := range []int{0, -1} {
		base := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, dayOffset)

		weekday := int(base.Weekday())
		if !containsInt(timer.Days, weekday) {
			continue
		}

		start := time.Date(base.Year(), base.Month(), base.Day(), startClock.Hour(), startClock.Minute(), 0, 0, now.Location())
		stop := time.Date(base.Year(), base.Month(), base.Day(), stopClock.Hour(), stopClock.Minute(), 0, 0, now.Location())
		if !stop.After(start) {
			stop = stop.Add(24 * time.Hour)
		}

		if (now.Equal(start) || now.After(start)) && now.Before(stop) {
			return true
		}
	}

	return false
}

func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
