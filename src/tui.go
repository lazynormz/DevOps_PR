package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	prGreen          = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	draftGray        = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // muted gray
	selectedStyle    = lipgloss.NewStyle().Bold(true)
	boxStyle         = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).Margin(0, 1)
	reviewerBox      = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).Margin(1, 1)
	reviewerName     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Width(20)
	requiredStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Width(9)
	voteStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Width(10)
	idStyle          = lipgloss.NewStyle().Faint(true).Width(24)
	sepStyle         = lipgloss.NewStyle().Faint(true)
	requiredRowStyle = lipgloss.NewStyle().Background(lipgloss.Color("8")).Bold(true)
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Align(lipgloss.Center).MarginBottom(1).Height(2)
)

func voteLabel(v int) string {
	switch v {
	case 10:
		return "Approved"
	case 5:
		return "Suggest"
	case 0:
		return "No Vote"
	case -5:
		return "Waiting"
	case -10:
		return "Rejected"
	default:
		return fmt.Sprintf("%d", v)
	}
}

type tuiModel struct {
	prs             []PullRequestInfo
	selected        int
	showDrafts      bool
	showMine        bool
	showNotReviewer bool
	userID          string
	width           int
	height          int
}

func (m tuiModel) filteredPRs() []PullRequestInfo {
	var filteredPRs []PullRequestInfo
	seenPRs := make(map[int]bool)

	if m.showNotReviewer {
		for _, pullRequest := range m.prs {
			if !m.showDrafts && pullRequest.IsDraft {
				continue
			}
			if !m.showMine && pullRequest.creatorID == m.userID {
				continue
			}
			if !seenPRs[pullRequest.id] {
				filteredPRs = append(filteredPRs, pullRequest)
				seenPRs[pullRequest.id] = true
			}
		}
		return filteredPRs
	}

	for _, pullRequest := range m.prs {
		if !m.showDrafts && pullRequest.IsDraft {
			continue
		}
		if pullRequest.creatorID == m.userID && !m.showMine {
			continue
		}
		isCurrentUserReviewer := false
		for _, reviewer := range pullRequest.reviewers {
			if reviewer.id == m.userID {
				isCurrentUserReviewer = true
				break
			}
		}
		if isCurrentUserReviewer && !seenPRs[pullRequest.id] {
			filteredPRs = append(filteredPRs, pullRequest)
			seenPRs[pullRequest.id] = true
		}
	}

	if m.showMine {
		for _, pullRequest := range m.prs {
			if pullRequest.creatorID == m.userID && !pullRequest.IsDraft && !seenPRs[pullRequest.id] {
				filteredPRs = append(filteredPRs, pullRequest)
				seenPRs[pullRequest.id] = true
			}
			if m.showDrafts && pullRequest.creatorID == m.userID && pullRequest.IsDraft && !seenPRs[pullRequest.id] {
				filteredPRs = append(filteredPRs, pullRequest)
				seenPRs[pullRequest.id] = true
			}
		}
	}
	return filteredPRs
}

func (m tuiModel) Init() tea.Cmd {
	return tea.EnterAltScreen
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
		case "r":
			m.showNotReviewer = !m.showNotReviewer
			m.selected = 0
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
	mainArea := ""
	if len(prs) == 0 {
		msg := "No open pull requests where you are set as a reviewer."
		if m.showNotReviewer {
			msg = "No open pull requests where you are NOT set as a reviewer."
		}
		mainArea = lipgloss.NewStyle().Width(m.width).Height(m.height-2).Align(lipgloss.Center, lipgloss.Center).Render(msg)
	} else {
		// Calculate max width for the PR line inside the box
		frameWidth, _ := boxStyle.GetFrameSize()
		maxBoxWidth := m.width / 2
		usableWidth := maxBoxWidth - frameWidth
		prLines := make([]string, len(prs))
		for i, pr := range prs {
			cursor := " "
			if i == m.selected {
				cursor = ">"
			}
			mode := ""
			if pr.IsDraft {
				mode = "[Draft] "
			}
			idStr := fmt.Sprintf("[%d]", pr.id)
			creatorStr := fmt.Sprintf("(by %s)", pr.creator)
			staticLen := len(cursor) + 1 + len(idStr) + 1 + len(mode) + 1 + len(creatorStr) + 1 // spaces between
			maxTitleLen := usableWidth - staticLen
			title := pr.title
			if maxTitleLen <= 0 {
				title = ""
			} else if len(title) > maxTitleLen {
				title = title[:maxTitleLen-3] + "..."
			}
			prLine := fmt.Sprintf("%s %s %s%s %s", cursor, idStr, mode, title, creatorStr)
			if len(prLine) > usableWidth {
				// Cut off from the right, but always keep ID and creator
				cutLen := usableWidth - len(creatorStr) - 1 // space before creatorStr
				if cutLen > 0 {
					prLine = prLine[:cutLen] + " " + creatorStr
				} else {
					prLine = cursor + " " + idStr + " " + creatorStr
				}
			}
			if pr.IsDraft {
				prLine = draftGray.Render(prLine)
			} else {
				prLine = prGreen.Render(prLine)
			}
			if i == m.selected {
				prLine = selectedStyle.Render(prLine)
			}
			prLines[i] = prLine
		}
		prBox := boxStyle.Width(maxBoxWidth).Align(lipgloss.Left).Render(lipgloss.JoinVertical(lipgloss.Left, prLines...))
		mainArea += lipgloss.Place(m.width, m.height/2, lipgloss.Center, lipgloss.Center, prBox)
		selectedPR := prs[m.selected]
		// Title area (big, centered)
		titleArea := titleStyle.Render(selectedPR.title)
		// Reviewer table area
		reviewerLines := []string{"Reviewers:",
			sepStyle.Render("┌" + strings.Repeat("─", 20) + "┬" + strings.Repeat("─", 9) + "┬" + strings.Repeat("─", 10) + "┬" + strings.Repeat("─", 24) + "┐"),
			"│" + reviewerName.Render("Name") + "│" + requiredStyle.Render("Required") + "│" + voteStyle.Render("Vote") + "│" + idStyle.Render("ID") + "│",
			sepStyle.Render("├" + strings.Repeat("─", 20) + "┼" + strings.Repeat("─", 9) + "┼" + strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 24) + "┤"),
		}
		ignoreIDs := map[string]bool{
			"1809cf47-1683-62b4-ab66-9dbfd3d291d6": true,
			"59e23168-dd18-4b40-9065-f3182d63ff1a": true,
		}
		for _, rev := range selectedPR.reviewers {
			if ignoreIDs[rev.id] {
				continue
			}
			nameStr := rev.displayName
			maxNameLen := 20
			if len(nameStr) > maxNameLen {
				nameStr = nameStr[:maxNameLen-3] + "..."
			}
			name := reviewerName.Render(nameStr)
			required := requiredStyle.Render("")
			if rev.isRequired {
				required = requiredStyle.Foreground(lipgloss.Color("2")).Render("✔ Yes")
			}
			vote := voteStyle.Render(voteLabel(rev.vote))
			idStr := rev.id
			maxIdLen := 24
			if len(idStr) > maxIdLen {
				idStr = idStr[:maxIdLen-3] + "..."
			}
			id := idStyle.Render(idStr)
			row := "│" + name + "│" + required + "│" + vote + "│" + id + "│"
			reviewerLines = append(reviewerLines, row)
		}
		reviewerLines = append(reviewerLines, sepStyle.Render("└"+strings.Repeat("─", 20)+"┴"+strings.Repeat("─", 9)+"┴"+strings.Repeat("─", 10)+"┴"+strings.Repeat("─", 24)+"┘"))
		// Combine title and reviewers in the box
		reviewerBoxStr := reviewerBox.Width(maxBoxWidth).Align(lipgloss.Left).Render(
			titleArea + "\n" + lipgloss.JoinVertical(lipgloss.Left, reviewerLines...))
		mainArea += "\n" + lipgloss.Place(m.width, m.height/2, lipgloss.Center, lipgloss.Top, reviewerBoxStr)
	}
	// Instructions at the bottom
	menuBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 2)
	// Styles for ON/OFF
	onStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)  // green
	offStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true) // red

	dKey := offStyle.Render("d")
	if m.showDrafts {
		dKey = onStyle.Render("d")
	}
	mKey := offStyle.Render("m")
	if m.showMine {
		mKey = onStyle.Render("m")
	}
	rKey := offStyle.Render("r")
	if m.showNotReviewer {
		rKey = onStyle.Render("r")
	}

	instructions := fmt.Sprintf("  ↑/↓ to navigate | %s: toggle drafts | %s: show/hide your own PRs | %s: show PRs where you are NOT a reviewer | q: quit  ", dKey, mKey, rKey)
	if m.width > 0 {
		instructions = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, menuBox.Render(instructions))
	}
	return mainArea + "\n" + instructions
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

func initialModel(prs []PullRequestInfo, userID string) tuiModel {
	return tuiModel{
		prs:             prs,
		selected:        0,
		showDrafts:      false,
		showMine:        false,
		showNotReviewer: false,
		userID:          userID,
		width:           0,
		height:          0,
	}
}

func RunTUI(prs []PullRequestInfo, userID string) {
	p := tea.NewProgram(initialModel(prs, userID))
	_ = p.Start()
}
