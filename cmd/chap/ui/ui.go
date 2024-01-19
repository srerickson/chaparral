package ui

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	client "github.com/srerickson/chaparral/client"
	"github.com/srerickson/chaparral/internal/diff"
	"golang.org/x/exp/maps"
)

var ErrCancled = errors.New("canceled by user input")

func CommitSummary(commit *client.Commit) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "%s %s\n",
		nameStyle.Render("preparing new object version:"),
		valueStyle.Render(fmt.Sprintf("%s v%d", commit.ObjectID, commit.Version)),
	)
	fmt.Fprintf(b, "%s: %s\n",
		nameStyle.Render("storage root"),
		valueStyle.Render(commit.StorageRootID))
	return b.String()
}

func PrettyDiff(r diff.Result) string {
	b := &strings.Builder{}
	for _, n := range r.Added {
		fmt.Fprintln(b, " +", addStyle.Render(n))
	}

	for _, n := range r.Modified {
		fmt.Fprintln(b, " ~", modStyle.Render(n))
	}
	for _, n := range r.Removed {
		fmt.Fprintln(b, " -", rmStyle.Render(n))
	}
	moved := maps.Keys(r.Renamed)
	sort.Strings(moved)
	for _, n := range moved {
		fmt.Fprintln(b, "  ", mvSrcStyle.Render(n), "->", mvDstStyle.Render(r.Renamed[n]))
	}
	return b.String() + "\n"
}

// Run Bubbletea program for uploading files
func RunUpload(cli *client.Client, up *client.Uploader, files []string) error {
	if len(files) == 0 {
		return nil
	}
	m := uploadUI{model: NewUploader(cli, up, files)}
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return err
	}
	m = final.(uploadUI)
	if m.canceled || m.model.Canceled {
		return ErrCancled
	}
	return m.model.Err
}

type uploadUI struct {
	model    Uploader
	canceled bool
}

func (m uploadUI) View() string  { return m.model.View() }
func (m uploadUI) Init() tea.Cmd { return m.model.UploadFiles() }
func (m uploadUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UploadDoneMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.canceled = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

// Run Bubbletea program for building stage from a directory
func RunStageDir(dir string, alg string, opts ...client.StageOption) (*client.Stage, error) {
	m := stageDirUI{model: NewStager(dir, alg, opts...)}
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}
	m = final.(stageDirUI)
	if m.canceled {
		return nil, ErrCancled
	}
	return m.model.Stage, m.model.Err
}

type stageDirUI struct {
	model    Stager
	canceled bool
}

func (m stageDirUI) Init() tea.Cmd { return m.model.StageDir() }
func (m stageDirUI) View() string  { return m.model.View() }
func (m stageDirUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case StagerDoneMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.canceled = true
			return m, tea.Quit
		}
	}
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

// Run Bubbletea program for confirming commit message
func RunConfirmCommit(commit *client.Commit) bool {
	m := commutUI{
		model: NewCommit(commit.User.Name, commit.User.Address, commit.Message),
	}
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return false
	}
	m = final.(commutUI)
	if m.canceled {
		return false
	}
	commit.User.Name = m.result.Name
	commit.User.Address = m.result.Email
	commit.Message = m.result.Message
	return m.result.Confirm
}

type commutUI struct {
	model    CommitMeta
	result   CommitResult
	canceled bool
}

func (m commutUI) Init() tea.Cmd { return CommitFocus() }
func (m commutUI) View() string  { return m.model.View() }
func (m commutUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case CommitResult:
		m.result = msg
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.canceled = true
			return m, tea.Quit
		}
	}
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

// Run Bubbletea program for uploading files
func RunDownload(cli *client.Client, rootID, object string, files map[string]string, replace bool) error {
	if len(files) == 0 {
		return nil
	}
	m := downloadUI{model: NewDownloader(cli, rootID, object, files, replace)}
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return err
	}
	m = final.(downloadUI)
	if m.canceled || m.model.Canceled {
		return ErrCancled
	}
	return m.model.Err
}

type downloadUI struct {
	model    Downloader
	canceled bool
}

func (m downloadUI) View() string  { return m.model.View() }
func (m downloadUI) Init() tea.Cmd { return m.model.DownloadFiles() }
func (m downloadUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DownloadDoneMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.canceled = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

func scale[T int64 | float64](byteSize T) (scaled float64, unit string) {
	var units = []string{"Bytes", "KB", "MB", "GB", "TB"}
	scaled = float64(byteSize)
	for i := 0; i < len(units); i++ {
		unit = units[i]
		if scaled < 1000 {
			return
		}
		scaled = scaled / 1000
	}
	return
}
