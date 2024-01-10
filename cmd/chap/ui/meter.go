package ui

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Meter struct {
	FileCount  int       // total number of files read through EOF
	TotalBytes int64     // total number of bytes read across all files
	Start      time.Time // time of first read/reset msg
	Elapsed    time.Duration
	Err        error

	msgChan chan meterMsg
}

func NewMeter() Meter {
	return Meter{
		msgChan: make(chan meterMsg, 64),
	}
}

func (m Meter) Update(msg tea.Msg) (Meter, tea.Cmd) {
	switch msg := msg.(type) {
	case meterResetMsg:
		m.reset()
		return m, nil
	case meterMsg:
		if errors.Is(msg.Err, io.EOF) {
			m.FileCount++
		}
		now := time.Now()
		if m.TotalBytes == 0 {
			m.Start = now
		}
		m.Elapsed = now.Sub(m.Start)
		if m.Elapsed == 0 {
			m.Elapsed = time.Microsecond * 1
		}
		m.TotalBytes += int64(msg.Size)
		return m, m.block() // wait for next meterMsg
	case meterFinalMsg:
		for _, err := range msg.Errs {
			if errors.Is(err, io.EOF) {
				m.FileCount++
			}
			m.TotalBytes += msg.Size
		}
		now := time.Now()
		if m.TotalBytes == 0 {
			m.Start = now
		}
		m.Elapsed = now.Sub(m.Start)
		if m.Elapsed == 0 {
			m.Elapsed = time.Microsecond * 1
		}
		m.TotalBytes += int64(msg.Size)
		return m, meterDone
	}
	return m, nil
}

func (m Meter) View() string {
	builder := strings.Builder{}
	elapsed := m.Elapsed.Seconds()
	rateScaled, rateUnit := scale(float64(m.TotalBytes) / elapsed)
	progScaled, progUnits := scale(m.TotalBytes)
	// number files
	builder.WriteString(nameStyle.Render("files:"))
	builder.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.FileCount)))
	builder.WriteString(" ")
	// bytes read
	builder.WriteString(nameStyle.Render("size:"))
	builder.WriteString(valueStyle.Render(fmt.Sprintf("%0.2f%s", progScaled, progUnits)))
	builder.WriteString(" ")
	// rate
	builder.WriteString(nameStyle.Render("rate:"))
	builder.WriteString(valueStyle.Render(fmt.Sprintf("%0.2f%s/s", rateScaled, rateUnit)))
	return builder.String()
}

type meterMsg struct {
	Size int
	Err  error
}

func (m Meter) block() tea.Cmd {
	return func() tea.Msg {
		return <-m.msgChan
	}
}

type meterFinalMsg struct {
	Size int64
	Errs []error
}

func (m Meter) wait() tea.Cmd {
	return func() tea.Msg {
		var final meterFinalMsg
		for m := range m.msgChan {
			final.Size += int64(m.Size)
			if m.Err != nil {
				final.Errs = append(final.Errs, m.Err)
			}
		}
		return final
	}
}

type meterResetMsg struct{}

func meterReset() tea.Msg {
	return meterResetMsg{}
}

func (m *Meter) reset() {
	m.Start = time.Time{}
	m.TotalBytes = 0
	m.FileCount = 0
}

type meterDoneMsg struct{}

func meterDone() tea.Msg {
	return meterDoneMsg{}
}
