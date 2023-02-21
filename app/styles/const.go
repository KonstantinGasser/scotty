package styles

const (
	// space must be reserved for the status either at
	// the top or the bottom however must be taken in
	// account for the other views
	heightStatusBar = 2

	// positioned at the bottom of the application
	// height of the input model for commands
	heightInputBar = 1
)

func AvailableHeight(height int) int {
	return height - heightStatusBar - heightInputBar
}
