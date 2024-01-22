package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	client "github.com/srerickson/chaparral/client"
	"golang.org/x/sync/errgroup"
)

type downloaderState int

const (
	downloaderInactive downloaderState = iota
	downloaderActive
	downloaderDone
)

var (
	FileMode, DirMode fs.FileMode = 0664, 0775
)

// UploadDoneMsg is a message used to signal that the downloader is done
type DownloadDoneMsg struct{}

type Downloader struct {
	state downloaderState // ui state
	meter Meter           // download progress UI

	Concurrency int // number of simultaneous downloads

	cli      *client.Client
	rootID   string
	objectID string

	// content path to local path
	files   map[string]string
	replace bool

	// result
	Err      error
	Skipped  []string // any skipped files
	Canceled bool
}

func NewDownloader(cli *client.Client, root, object string, files map[string]string, replace bool) Downloader {
	return Downloader{
		Concurrency: runtime.NumCPU(),
		cli:         cli,
		rootID:      root,
		objectID:    object,
		files:       files,
		replace:     replace,
	}
}

func (m Downloader) Update(msg tea.Msg) (Downloader, tea.Cmd) {
	switch msg := msg.(type) {
	case startDownloadMsg:
		m.meter = NewMeter()
		m.state = downloaderActive
		return m, m.doDownload()
	case downloaderResult:
		m.Err = msg.Err
		m.Skipped = msg.Skipped
		m.state = downloaderDone
		close(m.meter.msgChan)
		return m, m.meter.wait()
	case meterDoneMsg:
		return m, downloaderDoneCmd
	}
	var cmd tea.Cmd
	m.meter, cmd = m.meter.Update(msg)
	return m, cmd
}

func (m Downloader) View() string {
	b := &strings.Builder{}
	fmt.Fprintln(b, nameStyle.Render("downloading"), len(m.files), m.meter.View())
	for _, n := range m.Skipped {
		fmt.Fprintln(b, "skipped replacement:", n)
	}
	return b.String()
}

type startDownloadMsg struct{}

func (m Downloader) DownloadFiles() tea.Cmd {
	return func() tea.Msg {
		return startDownloadMsg{}
	}
}

func (m Downloader) doDownload() tea.Cmd {
	fn := func() tea.Msg {
		// list of files skipped
		skipped := []string{}
		skippedCh := make(chan string)
		skippedChWait := make(chan struct{})
		go func() {
			defer close(skippedChWait)
			for name := range skippedCh {
				skipped = append(skipped, name)
			}
		}()
		grp, ctx := errgroup.WithContext(context.Background())
		grp.SetLimit(1)
		for dst, src := range m.files {
			s, d := src, dst // loop variable captur
			grp.Go(func() error {
				err := m.saveContent(ctx, s, d, m.replace)
				if err != nil && errors.Is(err, errSkippedExisting) {
					skippedCh <- d
					err = nil
				}
				return err
			})
		}
		err := grp.Wait()
		close(skippedCh)
		<-skippedChWait
		return downloaderResult{Err: err, Skipped: skipped}
	}
	return tea.Batch(fn, m.meter.block())
}

type downloaderResult struct {
	Skipped []string
	Err     error
}

func downloaderDoneCmd() tea.Msg { return DownloadDoneMsg{} }

var errSkippedExisting = errors.New("skipped")

func (m *Downloader) saveContent(ctx context.Context, digest, localPath string, replace bool) error {
	reader, err := m.cli.GetContent(ctx, m.rootID, m.objectID, digest, "")
	if err != nil {
		return err
	}
	defer reader.Close()
	perm := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !replace {
		perm = os.O_WRONLY | os.O_CREATE | os.O_EXCL
		if _, err := os.Stat(localPath); err == nil {
			return fmt.Errorf("%s: %w", localPath, errSkippedExisting)
		}
	}
	if err := os.MkdirAll(filepath.Dir(localPath), DirMode); err != nil {
		return err
	}
	f, err := os.OpenFile(localPath, perm, FileMode)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, newMeterRead(reader, m.meter.msgChan)); err != nil {
		return err
	}
	return f.Sync()
}
