package event

import (
	tea "github.com/charmbracelet/bubbletea"
)

type GlobalRefreshBuffer struct{}

func RequestGlobalBufferRefresh() tea.Msg {
	return GlobalRefreshBuffer{}
}

// type BlockKeys []string
//
// func BlockKeysRequest(keys ...string) tea.Cmd {
// 	return func() tea.Msg {
// 		return BlockKeys(keys)
// 	}
// }
//
// type ReleaseKeys struct{}
//
// func ReleaseKeysRequest() tea.Cmd {
// 	return func() tea.Msg {
// 		return ReleaseKeys{}
// 	}
// }
//
// type Increment string
//
// func IncrementRequest(label string) tea.Cmd {
// 	return func() tea.Msg {
// 		return Increment(label)
// 	}
// }
//
// type TaillingPaused struct{}
//
// func TaillingPausedRequest() tea.Cmd {
// 	return func() tea.Msg {
// 		return TaillingPaused{}
// 	}
// }
//
// type TaillingResumed struct{}
//
// func TaillingResumedRequest() tea.Cmd {
// 	return func() tea.Msg {
// 		return TaillingResumed{}
// 	}
// }
