package styles

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	ColorLogo              = lipgloss.Color("93")
	ColorBorder            = lipgloss.Color("0")
	ColorBorderActive      = ColorLogo // tight to the logo color
	ColorStatusBarLogCount = lipgloss.Color("93")
	ColorStatusBarBeamInfo = lipgloss.Color("22")
	ColorErrorBackground   = lipgloss.Color("160")
)

func RandColor() lipgloss.Color {
	// including extended ANSI Grayscale color reach from 0-255
	// see https://github.com/muesli/termenv which is used for lipgloss
	return lipgloss.Color(fmt.Sprint(rand.Intn(256)))
}

func init() {
	rand.Seed(time.Now().Unix())
}
