package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	client "github.com/srerickson/chaparral/client"
)

type Stager struct {
	Stage *client.Stage
	Err   error
	dir   string
	alg   string
	opts  []client.StageOption
	meter Meter
}

type StagerDoneMsg struct{}

func NewStager(dir string, alg string, opts ...client.StageOption) Stager {
	return Stager{
		dir:   dir,
		alg:   alg,
		opts:  opts,
		meter: NewMeter(),
	}
}

func (m Stager) Update(msg tea.Msg) (Stager, tea.Cmd) {
	switch msg := msg.(type) {
	case stagerResult:
		m.Err = msg.Err
		m.Stage = msg.Stage
		close(m.meter.msgChan)
		return m, m.meter.wait()
	case meterDoneMsg:
		return m, stagerDoneCmd
	}
	var cmd tea.Cmd
	m.meter, cmd = m.meter.Update(msg)
	return m, cmd
}

func (m Stager) View() string {
	b := &strings.Builder{}
	fmt.Fprintln(b, nameStyle.Render("source directory:"), valueStyle.Render(m.dir))
	fmt.Fprintln(b, nameStyle.Render("digest alg:"), valueStyle.Render(m.alg), m.meter.View())
	return b.String()
}

func (m Stager) StageDir() tea.Cmd {
	stageFn := func() tea.Msg {
		onread := client.OnRead(func(size int, err error) {
			m.meter.msgChan <- meterMsg{Size: size, Err: err}
		})
		opts := append(m.opts, onread, client.AddDir(m.dir))
		st, err := client.NewStage(m.alg, opts...)
		return stagerResult{
			Stage: st,
			Err:   err,
		}
	}
	return tea.Sequence(meterReset, tea.Batch(stageFn, m.meter.block()))
}

type stagerResult struct {
	Stage *client.Stage
	Err   error
}

func stagerDoneCmd() tea.Msg {
	return StagerDoneMsg{}

}
