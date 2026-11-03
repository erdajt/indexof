package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/erdajt/indexof/internal/netclient"
)

type State int

const (
	StateQueued State = iota
	StateRunning
	StatePaused
	StateDone
	StateError
	StateCancelled
)

func (s State) String() string {
	switch s {
	case StateQueued:
		return "queued"
	case StateRunning:
		return "downloading"
	case StatePaused:
		return "paused"
	case StateDone:
		return "done"
	case StateError:
		return "error"
	case StateCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

type Job struct {
	ID        string
	URL       string
	Name      string
	Dest      string
	Total     int64
	Written   int64
	State     State
	Err       string
	OpenAfter bool

	cancel context.CancelFunc
	mu     sync.Mutex
}

type Manager struct {
	client *netclient.Client
	jobs   []*Job
	mu     sync.Mutex
}

func NewManager(client *netclient.Client) *Manager {
	return &Manager{client: client}
}

func (m *Manager) Jobs() []*Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*Job, len(m.jobs))
	copy(out, m.jobs)
	return out
}

func (m *Manager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, j := range m.jobs {
		j.mu.Lock()
		if j.State == StateRunning || j.State == StateQueued {
			n++
		}
		j.mu.Unlock()
	}
	return n
}

func (m *Manager) Start(url, name, dest string, size int64, openAfter bool) *Job {
	job := &Job{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		URL:       url,
		Name:      name,
		Dest:      dest,
		Total:     size,
		State:     StateQueued,
		OpenAfter: openAfter,
	}

	m.mu.Lock()
	m.jobs = append(m.jobs, job)
	m.mu.Unlock()

	m.run(job)
	return job
}

func (j *Job) ClearOpenAfter() {
	j.mu.Lock()
	j.OpenAfter = false
	j.mu.Unlock()
}

func (m *Manager) run(job *Job) {
	ctx, cancel := context.WithCancel(context.Background())
	job.mu.Lock()
	job.cancel = cancel
	job.State = StateRunning
	job.mu.Unlock()

	go func() {
		var existing int64
		if st, err := os.Stat(job.Dest); err == nil {
			existing = st.Size()
		}

		written, err := m.client.Download(ctx, job.URL, job.Dest, existing, func(w, t int64) {
			job.mu.Lock()
			job.Written = w
			if t > 0 {
				job.Total = t
			}
			job.mu.Unlock()
		})

		job.mu.Lock()
		defer job.mu.Unlock()
		job.Written = written

		if ctx.Err() != nil {
			if job.State != StatePaused {
				job.State = StateCancelled
			}
			return
		}
		if err != nil {
			job.State = StateError
			job.Err = err.Error()
			return
		}
		job.State = StateDone
	}()
}

func (j *Job) Snapshot() (state State, written, total int64, name, errStr string, openAfter bool, dest string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.State, j.Written, j.Total, j.Name, j.Err, j.OpenAfter, j.Dest
}

func (j *Job) GetState() State {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.State
}

func (m *Manager) Pause(job *Job) {
	job.mu.Lock()
	defer job.mu.Unlock()
	if job.State == StateRunning && job.cancel != nil {
		job.cancel()
		job.State = StatePaused
	}
}

func (m *Manager) Resume(job *Job) {
	job.mu.Lock()
	st := job.State
	job.mu.Unlock()
	if st == StatePaused {
		m.run(job)
	}
}

func (m *Manager) Cancel(job *Job) {
	job.mu.Lock()
	defer job.mu.Unlock()
	if job.cancel != nil {
		job.cancel()
	}
	job.State = StateCancelled
}

func (m *Manager) ClearFinished() {
	m.mu.Lock()
	defer m.mu.Unlock()
	filtered := m.jobs[:0]
	for _, j := range m.jobs {
		j.mu.Lock()
		st := j.State
		j.mu.Unlock()
		if st != StateDone && st != StateError && st != StateCancelled {
			filtered = append(filtered, j)
		}
	}
	m.jobs = filtered
}

func (m *Manager) DestPath(downloadDir string, moveToCwd bool, name string) string {
	if moveToCwd {
		cwd, err := os.Getwd()
		if err == nil {
			return filepath.Join(cwd, name)
		}
	}
	return filepath.Join(downloadDir, name)
}
