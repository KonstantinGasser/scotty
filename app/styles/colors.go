package styles

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v2"
)

type Color struct {
	Border    lipgloss.Color `yaml:"border"`
	Error     lipgloss.Color `yaml:"error"`
	Highlight lipgloss.Color `yaml:"highlight"`
	Light     lipgloss.Color `yaml:"light"`
}

type Config struct {
	Colors Color `yaml:"colors"`
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

func LoadConfig() (*Config, error) {
	var cfg Config

	yamlFile, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	defer yamlFile.Close()

	yamlBytes, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(yamlBytes, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func init() {
	rand.Seed(time.Now().Unix())

	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("error loading config file: %v\n", err)
		return
	}

	DefaultColor = config.Colors
}
