package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	prGreen       = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	draftGray     = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // muted gray
	selectedStyle = lipgloss.NewStyle().Bold(true)
)

type tuiModel struct {
	prs        []PullRequestInfo
	selected   int
	showDrafts bool
	showMine   bool
	userID     string
}

func initialModel(prs []PullRequestInfo, userID string) tuiModel {
	// By default, hide user's own PRs (showMine = true)
	return tuiModel{prs: prs, selected: 0, showDrafts: false, showMine: true, userID: userID}
}

func (m tuiModel) filteredPRs() []PullRequestInfo {
	filtered := m.prs
	if !m.showDrafts {
		var tmp []PullRequestInfo
		for _, pr := range filtered {
			if !pr.IsDraft {
				tmp = append(tmp, pr)
			}
		}
		filtered = tmp
	}
	if m.showMine {
		// Hide my own PRs, show only PRs created by others
		var others []PullRequestInfo
		for _, pr := range filtered {
			if pr.creatorID != m.userID {
				others = append(others, pr)
			}
		}
		filtered = others
	}
	return filtered
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.filteredPRs())-1 {
				m.selected++
			}
		case "d":
			m.showDrafts = !m.showDrafts
			m.selected = 0
		case "m":
			m.showMine = !m.showMine
			m.selected = 0
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m tuiModel) View() string {
	prs := m.filteredPRs()
	// Check if all non-draft PRs are mine
	nonDraftCount := 0
	mineCount := 0
	for _, pr := range m.prs {
		if !pr.IsDraft {
			nonDraftCount++
			if pr.creatorID == m.userID {
				mineCount++
			}
		}
	}
	view := "Open Pull Requests (↑/↓ to navigate, d to toggle drafts, m to toggle mine, q to quit)\n\n"
	if nonDraftCount > 0 && nonDraftCount == mineCount {
		view += "No new pull requests. All active PRs are yours.\n\n"
	}
	if len(prs) == 0 {
		return view + "No open pull requests. Press q to quit."
	}
	for i, pr := range prs {
		cursor := " "
		if i == m.selected {
			cursor = ">"
		}
		mode := ""
		var prLine string
		if pr.IsDraft {
			mode = "[Draft] "
			prLine = draftGray.Render(fmt.Sprintf("%s [%d] %s%s (by %s)", cursor, pr.id, mode, pr.title, pr.creator))
		} else {
			prLine = prGreen.Render(fmt.Sprintf("%s [%d] %s%s (by %s)", cursor, pr.id, mode, pr.title, pr.creator))
		}
		if i == m.selected {
			prLine = selectedStyle.Render(prLine)
		}
		view += prLine + "\n"
	}
	view += "\n"
	selectedPR := prs[m.selected]
	view += fmt.Sprintf("Selected PR: %s\nReviewers:\n", selectedPR.title)
	for _, rev := range selectedPR.reviewers {
		view += fmt.Sprintf(" - %s (ID: %s, Required: %t, Vote: %d)\n", rev.displayName, rev.id, rev.isRequired, rev.vote)
	}
	return view
}

// RunTUIWithError displays an error message in the TUI and exits on key press
func RunTUIWithError(prs []PullRequestInfo, errorMsg string) {
	errModel := errorTUIModel{errorMsg: errorMsg}
	p := tea.NewProgram(errModel)
	_ = p.Start()
}

type errorTUIModel struct {
	errorMsg string
}

func (m errorTUIModel) Init() tea.Cmd { return nil }

func (m errorTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m errorTUIModel) View() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render(
		"Error: " + m.errorMsg + "\nPress any key to exit.")
}

func RunTUI(prs []PullRequestInfo, userID string) {
	p := tea.NewProgram(initialModel(prs, userID))
	_ = p.Start()
}
