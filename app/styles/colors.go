package styles

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	ColorLogo              = lipgloss.Color("93")
	ColorBorder            = lipgloss.Color("#9F2DEB")
	ColorBorderActive      = ColorLogo // tight to the logo color
	ColorStatusBarLogCount = lipgloss.Color("#FF4C94")
	ColorStatusBarBeamInfo = lipgloss.Color("22")
	ColorErrorBackground   = lipgloss.Color("160")
	ColorError             = lipgloss.Color("#FF0000")
)

type Color struct {
	Border    lipgloss.Color
	Error     lipgloss.Color
	Highlight lipgloss.Color
	Light     lipgloss.Color
}

var DefaultColor = Color{
	Border:    lipgloss.Color("97"),
	Error:     lipgloss.Color("31"),
	Light:     lipgloss.Color("7"),
	Highlight: lipgloss.Color("11"),
}

func RandColor() (lipgloss.Color, lipgloss.Color) {
	// including extended ANSI Grayscale color reach from 0-255
	// see https://github.com/muesli/termenv which is used for lipgloss

	color := lipgloss.Color(fmt.Sprint(rand.Intn(256)))

	return color, InverseColor(color)
}

func brightnessRation(r, g, b uint32) int {

	r = uint32(math.Pow(float64(r/255), 2.2))
	g = uint32(math.Pow(float64(g/255), 2.2))
	b = uint32(math.Pow(float64(b/255), 2.2))

	delta := float64(0.2126)*float64(r) + float64(0.07151)*float64(g) + float64(0.0721)*float64(b)

	if delta < 0.5 {
		return 0
	}

	return 1
}

func InverseColor(c lipgloss.Color) lipgloss.Color {

	r, g, b, _ := c.RGBA()

	var inverse = lipgloss.Color("15") // color white

	if (float64(r)*0.299)+(float64(g)*0.587)+(float64(b)*0.114) > 150 {
		inverse = lipgloss.Color("0") // color black
	}

	return inverse
}

func init() {
	rand.Seed(time.Now().Unix())
}
