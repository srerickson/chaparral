package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	client "github.com/srerickson/chaparral/client"
	"golang.org/x/sync/errgroup"
)

type uploaderState int

const (
	// ui State
	uploaderInactive uploaderState = iota // no focus
	uploaderConfirming
	doingUpload
	uploaderDone
)

// UploadDoneMsg is a message used to signal that the uploader is done
type UploadDoneMsg struct{}

type Uploader struct {
	state   uploaderState   // ui state
	meter   Meter           // upload progress UI
	confirm textinput.Model // confirm upload input

	// configs for uploading
	SkipConfirm bool // don't require user confirmation
	Concurrency int  // number of simultaneous uploads

	// for upload
	cli   *client.Client
	up    *client.Uploader
	files []string // files to upload

	// result
	Err      error
	Canceled bool
}

func NewUploader(cli *client.Client, up *client.Uploader, files []string) Uploader {
	confirm := textinput.New()
	confirm.Cursor.Style = cursorStyle
	confirm.PlaceholderStyle = faintText
	return Uploader{
		confirm:     confirm,
		Concurrency: runtime.NumCPU(),
		cli:         cli,
		up:          up,
		files:       files,
	}
}

func (m Uploader) Update(msg tea.Msg) (Uploader, tea.Cmd) {
	switch msg := msg.(type) {
	case prepareUploadMsg:
		m.meter.reset()
		if m.SkipConfirm {
			m.state = doingUpload
			m.meter = NewMeter()
			return m, m.doUpload()
		}
		// setup confirm ui
		m.confirm.Prompt = "Upload? "
		m.confirm.Placeholder = "yes/no"
		m.state = uploaderConfirming
		m.confirm.PromptStyle = promptFocusStyle
		return m, tea.Batch(m.confirm.Focus(), textinput.Blink)
	case tea.KeyMsg:
		if msg.String() == "enter" && m.confirm.Focused() {
			switch m.confirm.Value() {
			case "yes", "y":
				m.confirm.Blur()
				m.confirm.SetValue("")
				m.meter = NewMeter()
				m.state = doingUpload
				return m, m.doUpload()
			case "no", "n":
				m.confirm.Blur()
				m.confirm.SetValue("")
				m.Canceled = true
				return m, uploaderDoneCmd
			}
		}
	case uploaderResult:
		m.Err = msg.err
		m.state = uploaderDone
		close(m.meter.msgChan)
		return m, m.meter.wait()
	case meterDoneMsg:
		return m, uploaderDoneCmd
	}
	cmds := make([]tea.Cmd, 2)
	m.meter, cmds[0] = m.meter.Update(msg)
	m.confirm, cmds[1] = m.confirm.Update(msg)
	return m, tea.Batch(cmds...)
}

func (m Uploader) View() string {
	b := &strings.Builder{}
	num := len(m.files)
	end := "s"
	if num == 1 {
		end = ""
	}
	fmt.Fprintf(b, nameStyle.Render("file%s to upload")+": "+valueStyle.Render("%d")+"\n", end, len(m.files))
	switch m.state {
	case uploaderConfirming:
		b.WriteString(m.confirm.View())
	case doingUpload:
		b.WriteString(nameStyle.Render("uploading "))
		b.WriteString(m.meter.View())
	case uploaderDone:
		b.WriteString(nameStyle.Render("uploading "))
		b.WriteString(m.meter.View())
		b.WriteString(" done")
	}
	b.WriteString("\n")
	return b.String()
}

type prepareUploadMsg struct{}

func (m Uploader) UploadFiles() tea.Cmd {
	return func() tea.Msg {
		return prepareUploadMsg{}
	}
}

func (m Uploader) doUpload() tea.Cmd {
	uploadFn := func() tea.Msg {
		if m.Err != nil {
			return uploaderResult{err: m.Err}
		}
		grp, ctx := errgroup.WithContext(context.Background())
		grp.SetLimit(m.Concurrency)
		for _, n := range m.files {
			fileName := n
			grp.Go(func() error {
				f, err := os.Open(fileName)
				if err != nil {
					return err
				}
				_, err = m.cli.Upload(ctx, m.up.UploadPath, newMeterRead(f, m.meter.msgChan))
				if err != nil {
					return errors.Join(err, f.Close())
				}
				return f.Close()
			})
		}
		return uploaderResult{err: grp.Wait()}
	}
	return tea.Batch(uploadFn, m.meter.block())
}

// Upload UploadResult
type uploaderResult struct {
	err error
}

func uploaderDoneCmd() tea.Msg { return UploadDoneMsg{} }

type meterReader struct {
	base io.Reader
	ch   chan meterMsg
}

func newMeterRead(reader io.Reader, ch chan meterMsg) *meterReader {
	return &meterReader{base: reader, ch: ch}
}

func (r *meterReader) Read(p []byte) (int, error) {
	n, err := r.base.Read(p)
	r.ch <- meterMsg{Size: n, Err: err}
	return n, err
}
