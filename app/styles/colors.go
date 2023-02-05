package styles

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/lipgloss"
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
