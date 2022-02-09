package service

import (
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/mariiatuzovska/workers/models"
)

type Service interface {
	Query() chan *models.Message
	Size() int
	Clear()
	Start()
}

type service struct {
	size    int
	maxProc int
	tmp     int
	mux     sync.Mutex
	w       io.Writer
	storage []models.Messages
	query   chan *models.Message
	cdone   chan int
}

func NewService(size, maxProc int, w io.Writer) Service {
	if size < 1 {
		size = 1
	}

	maxProcRuntime := runtime.GOMAXPROCS(maxProc)
	if maxProc <= 0 {
		maxProc = maxProcRuntime
	}

	storage := make([]models.Messages, maxProc)
	for i := range storage {
		storage[i] = make(models.Messages, 0, size)
	}

	return &service{
		size:    size,
		maxProc: maxProc,
		tmp:     0,
		mux:     sync.Mutex{},
		w:       w,
		storage: storage,
		query:   make(chan *models.Message, 100000),
		cdone:   make(chan int),
	}
}

func (s *service) Query() chan *models.Message {
	return s.query
}

func (s *service) Clear() {
	s.mux.Lock()
	defer s.mux.Unlock()
	for i := range s.storage {
		s.create(s.storage[i])
		s.storage[i] = make(models.Messages, 0, s.size)
	}
}

func (s *service) Size() int {
	return s.size
}

func (s *service) Start() {
	s.schedule()
	for i := range s.storage {
		s.cdone <- i
	}
}

func (s *service) create(slice models.Messages) {
	s.tmp++
	for _, msg := range slice {
		fmt.Fprintf(s.w, "[%4d] | %s | %s | %s\n",
			s.tmp, msg.Tag, msg.CreatedAt.Format(time.RFC3339Nano), msg.Text)
	}
	fmt.Fprintln(s.w)
	time.Sleep(1 * time.Second)
}

func (s *service) done(i int) {
	s.cdone <- i
}

func (s *service) schedule() {
	go func() {
		for {
			i, opened := <-s.cdone
			if !opened {
				return
			}
			s.run(i)
		}
	}()
}

func (s *service) run(i int) {
	go func() {
		for {
			s.storage[i] = append(s.storage[i], <-s.query)
			if len(s.storage[i]) == s.size-1 {
				break
			}
		}
		// s.create(slice)
		s.storage[i] = make(models.Messages, 0, s.size)
		s.done(i)
	}()
}
