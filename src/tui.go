package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	prGreen       = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	draftGray     = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // muted gray
	selectedStyle = lipgloss.NewStyle().Bold(true)
	boxStyle      = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).Margin(0, 1)
	reviewerBox   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).Margin(1, 1)
	reviewerName  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Width(20)
	requiredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Width(9)
	voteStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Width(10)
	idStyle       = lipgloss.NewStyle().Faint(true).Width(24)
	sepStyle      = lipgloss.NewStyle().Faint(true)
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Align(lipgloss.Center).MarginBottom(1).Height(2)
)

func voteLabel(vote int) string {
	switch vote {
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
		return fmt.Sprintf("%d", vote)
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

func (model tuiModel) filteredPRs() []PullRequestInfo {
	var filteredPRs []PullRequestInfo
	seenPRs := make(map[int]bool)

	if model.showNotReviewer {
		for _, pullRequest := range model.prs {
			if !model.showDrafts && pullRequest.IsDraft {
				continue
			}
			if !model.showMine && pullRequest.creatorID == model.userID {
				continue
			}
			if !seenPRs[pullRequest.id] {
				filteredPRs = append(filteredPRs, pullRequest)
				seenPRs[pullRequest.id] = true
			}
		}
		return filteredPRs
	}

	for _, pullRequest := range model.prs {
		if !model.showDrafts && pullRequest.IsDraft {
			continue
		}
		if pullRequest.creatorID == model.userID && !model.showMine {
			continue
		}
		isCurrentUserReviewer := false
		for _, reviewer := range pullRequest.reviewers {
			if reviewer.id == model.userID {
				isCurrentUserReviewer = true
				break
			}
		}
		if isCurrentUserReviewer && !seenPRs[pullRequest.id] {
			filteredPRs = append(filteredPRs, pullRequest)
			seenPRs[pullRequest.id] = true
		}
	}

	if model.showMine {
		for _, pullRequest := range model.prs {
			if pullRequest.creatorID == model.userID && !pullRequest.IsDraft && !seenPRs[pullRequest.id] {
				filteredPRs = append(filteredPRs, pullRequest)
				seenPRs[pullRequest.id] = true
			}
			if model.showDrafts && pullRequest.creatorID == model.userID && pullRequest.IsDraft && !seenPRs[pullRequest.id] {
				filteredPRs = append(filteredPRs, pullRequest)
				seenPRs[pullRequest.id] = true
			}
		}
	}
	return filteredPRs
}

func (model tuiModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (model tuiModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.KeyMsg:
		switch message.String() {
		case "up", "k":
			if model.selected > 0 {
				model.selected--
			}
		case "down", "j":
			if model.selected < len(model.filteredPRs())-1 {
				model.selected++
			}
		case "d":
			model.showDrafts = !model.showDrafts
			model.selected = 0
		case "m":
			model.showMine = !model.showMine
			model.selected = 0
		case "r":
			model.showNotReviewer = !model.showNotReviewer
			model.selected = 0
		case "q", "esc", "ctrl+c":
			return model, tea.Quit
		}
	case tea.WindowSizeMsg:
		model.width = message.Width
		model.height = message.Height
	}
	return model, nil
}

func (model tuiModel) View() string {
	prs := model.filteredPRs()
	// Check if all non-draft PRs are mine
	nonDraftCount := 0
	mineCount := 0
	for _, pr := range model.prs {
		if !pr.IsDraft {
			nonDraftCount++
			if pr.creatorID == model.userID {
				mineCount++
			}
		}
	}
	mainArea := ""
	if len(prs) == 0 {
		message := "No open pull requests where you are set as a reviewer."
		if model.showNotReviewer {
			message = "No open pull requests where you are NOT set as a reviewer."
		}
		mainArea = lipgloss.NewStyle().Width(model.width).Height(model.height-2).Align(lipgloss.Center, lipgloss.Center).Render(message)
	} else {
		// Calculate max width for the PR line inside the box
		frameWidth, _ := boxStyle.GetFrameSize()
		maxBoxWidth := model.width / 2
		usableWidth := maxBoxWidth - frameWidth
		prLines := make([]string, len(prs))
		for index, pr := range prs {
			cursor := " "
			if index == model.selected {
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
			if index == model.selected {
				prLine = selectedStyle.Render(prLine)
			}
			prLines[index] = prLine
		}
		prBox := boxStyle.Width(maxBoxWidth).Align(lipgloss.Left).Render(lipgloss.JoinVertical(lipgloss.Left, prLines...))
		mainArea += lipgloss.Place(model.width, model.height/2, lipgloss.Center, lipgloss.Center, prBox)
		selectedPR := prs[model.selected]
		titleArea := titleStyle.Render(selectedPR.title)
		repoArea := lipgloss.NewStyle().Faint(true).Align(lipgloss.Center).Render(selectedPR.repositoryName)
		reviewerLines := []string{"Reviewers:",
			sepStyle.Render("┌" + strings.Repeat("─", 20) + "┬" + strings.Repeat("─", 9) + "┬" + strings.Repeat("─", 10) + "┬" + strings.Repeat("─", 24) + "┐"),
			"│" + reviewerName.Render("Name") + "│" + requiredStyle.Render("Required") + "│" + voteStyle.Render("Vote") + "│" + idStyle.Render("ID") + "│",
			sepStyle.Render("├" + strings.Repeat("─", 20) + "┼" + strings.Repeat("─", 9) + "┼" + strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 24) + "┤"),
		}
		ignoreIDs := map[string]bool{
			"1809cf47-1683-62b4-ab66-9dbfd3d291d6": true,
			"59e23168-dd18-4b40-9065-f3182d63ff1a": true,
		}
		for _, reviewer := range selectedPR.reviewers {
			if ignoreIDs[reviewer.id] {
				continue
			}
			nameStr := reviewer.displayName
			maxNameLen := 20
			if len(nameStr) > maxNameLen {
				nameStr = nameStr[:maxNameLen-3] + "..."
			}
			name := reviewerName.Render(nameStr)
			required := requiredStyle.Render("")
			if reviewer.isRequired {
				required = requiredStyle.Foreground(lipgloss.Color("2")).Render("✔ Yes")
			}
			vote := voteStyle.Render(voteLabel(reviewer.vote))
			idStr := reviewer.id
			maxIdLen := 24
			if len(idStr) > maxIdLen {
				idStr = idStr[:maxIdLen-3] + "..."
			}
			id := idStyle.Render(idStr)
			row := "│" + name + "│" + required + "│" + vote + "│" + id + "│"
			reviewerLines = append(reviewerLines, row)
		}
		reviewerLines = append(reviewerLines, sepStyle.Render("└"+strings.Repeat("─", 20)+"┴"+strings.Repeat("─", 9)+"┴"+strings.Repeat("─", 10)+"┴"+strings.Repeat("─", 24)+"┘"))

		reviewerBoxStr := reviewerBox.Width(maxBoxWidth).Align(lipgloss.Left).Render(
			titleArea + "\n" + repoArea + "\n" + lipgloss.JoinVertical(lipgloss.Left, reviewerLines...))
		mainArea += "\n" + lipgloss.Place(model.width, model.height/2, lipgloss.Center, lipgloss.Top, reviewerBoxStr)
	}

	menuBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 2)
	onStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)  // green
	offStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true) // red

	dKey := offStyle.Render("d")
	if model.showDrafts {
		dKey = onStyle.Render("d")
	}
	mKey := offStyle.Render("m")
	if model.showMine {
		mKey = onStyle.Render("m")
	}
	rKey := offStyle.Render("r")
	if model.showNotReviewer {
		rKey = onStyle.Render("r")
	}

	instructions := fmt.Sprintf("  ↑/↓ to navigate | %s: toggle drafts | %s: show/hide your own PRs | %s: show PRs where you are NOT a reviewer | q: quit  ", dKey, mKey, rKey)
	if model.width > 0 {
		instructions = lipgloss.PlaceHorizontal(model.width, lipgloss.Center, menuBox.Render(instructions))
	}
	return mainArea + "\n" + instructions
}

// RunTUIWithError displays an error message in the TUI and exits on key press
func RunTUIWithError(prs []PullRequestInfo, errorMsg string) {
	errModel := errorTUIModel{errorMsg: errorMsg}
	program := tea.NewProgram(errModel)
	program.Run()
}

type errorTUIModel struct {
	errorMsg string
}

func (model errorTUIModel) Init() tea.Cmd { return nil }

func (model errorTUIModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message.(type) {
	case tea.KeyMsg:
		return model, tea.Quit
	}
	return model, nil
}

func (model errorTUIModel) View() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render(
		"Error: " + model.errorMsg + "\nPress any key to exit.")
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
	program := tea.NewProgram(initialModel(prs, userID))
	program.Run()
}
