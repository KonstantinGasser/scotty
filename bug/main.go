package main

// Link against ncurses with wide character support in case goncurses doesn't

// #cgo !darwin,!freebsd,!openbsd pkg-config: ncursesw
// #cgo darwin freebsd openbsd LDFLAGS: -lncurses
// #include <stdlib.h>
// #include <locale.h>
// #include <sys/select.h>
// #include <sys/ioctl.h>
//
// static void grv_FD_ZERO(void *set) {
// 	FD_ZERO((fd_set *)set);
// }
//
// static void grv_FD_SET(int fd, void *set) {
// 	FD_SET(fd, (fd_set *)set);
// }
//
// static int grv_FD_ISSET(int fd, void *set) {
// 	return FD_ISSET(fd, (fd_set *)set);
// }
//
import "C"
import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func windowSize() (int, int, error) {
	var winSize C.struct_winsize

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), C.TIOCGWINSZ, uintptr(unsafe.Pointer(&winSize)))
	if err != 0 {
		return 0, 0, err
	}

	return int(winSize.ws_col), int(winSize.ws_row), nil
}

type model struct {
	view  viewport.Model
	style lipgloss.Style
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg: // just for the sake of easiness
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.view.Width = msg.Width
		m.view.Height = msg.Height
		m.view.SetContent(fmt.Sprintf(`I would expect the border to span around the viewport with the width of: %d.
		Checking the viewport.Model.Width gives this value: %d
		Both the expected width and the viewport.Model.Width match, however, 
		the border is clearly spanning less than the width
		
		HorizontalFrameSize: %d
		Height: %d`, msg.Width, m.view.Width, m.view.Style.GetHorizontalFrameSize(), msg.Height))
	}

	var cmd tea.Cmd
	m.view, cmd = m.view.Update(msg)

	return m, cmd
}

func (m *model) View() string {
	return m.style.Render(
		m.view.View(),
	)
}

func main() {

	width, height, _ := windowSize()

	vp := viewport.New(width, height)
	// vp.Style = lipgloss.NewStyle().
	// 	Border(lipgloss.RoundedBorder()).
	// 	BorderForeground(lipgloss.Color("24"))
	// Width(width)

	vp.SetContent(fmt.Sprintf(`I would expect the border to span around the viewport with the width of: %d.
Checking the viewport.Model.Width gives this value: %d
Both the expected width and the viewport.Model.Width match, however, 
the border is clearly spanning less than the width

HorizontalFrameSize: %d
Height: %d`, width, vp.Width, vp.Style.GetHorizontalFrameSize(), height))

	m := model{
		view: vp,
		style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("24")).Width(width - 2).Height(height - 40),
	}

	app := tea.NewProgram(&m, tea.WithAltScreen())
	if _, err := app.Run(); err != nil {
		panic(err)
	}

}
