package main

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	_ "github.com/guigui-gui/guigui/basicwidget/cjkfont"
	"github.com/masakurapa/cchist/internal/types"
)

type screen int

const (
	screenList screen = iota
	screenDetail
)

func getProjectDir() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dirName := strings.ReplaceAll(pwd, "/", "-")
	return filepath.Join(home, ".claude", "projects", dirName), nil
}

func loadSessions(projectDir string) ([]types.Session, error) {
	files, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, err
	}
	var sessions []types.Session
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".jsonl" {
			continue
		}
		s, err := types.ParseSession(filepath.Join(projectDir, f.Name()))
		if err != nil {
			continue
		}
		if s.ID == "" {
			s.ID = strings.TrimSuffix(f.Name(), ".jsonl")
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

type Root struct {
	guigui.DefaultWidget

	background basicwidget.Background

	// list screen
	sessionTable     basicwidget.Table[int]
	sessionTableRows []basicwidget.TableRow[int]

	// detail screen
	backButton  basicwidget.Button
	titleText   basicwidget.Text
	summaryText basicwidget.Text
	panel       basicwidget.Panel
	msgList     msgListWidget

	layoutItems []guigui.LinearLayoutItem
	headerItems []guigui.LinearLayoutItem

	projectDir    string
	sessions      []types.Session
	items         []types.TimelineItem
	selectedIdx   int
	currentScreen screen
}

func (r *Root) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&r.background)

	u := basicwidget.UnitSize(context)

	switch r.currentScreen {
	case screenList:
		adder.AddWidget(&r.sessionTable)

		r.sessionTable.SetItemHeight(u)
		r.sessionTable.SetColumns([]basicwidget.TableColumn{
			{HeaderText: "id", Width: guigui.FixedSize(u * 10)},
			{HeaderText: "name", Width: guigui.FlexibleSize(1)},
			{HeaderText: "updated", Width: guigui.FixedSize(u * 6)},
		})
		n := len(r.sessions)
		if len(r.sessionTableRows) < n {
			r.sessionTableRows = slices.Grow(r.sessionTableRows, n-len(r.sessionTableRows))[:n]
		} else {
			r.sessionTableRows = slices.Delete(r.sessionTableRows, n, len(r.sessionTableRows))
		}
		for i, s := range r.sessions {
			r.sessionTableRows[i].Value = i
			if len(r.sessionTableRows[i].Cells) < 3 {
				r.sessionTableRows[i].Cells = make([]basicwidget.TableCell, 3)
			}
			r.sessionTableRows[i].Cells[0].Text = s.ID
			r.sessionTableRows[i].Cells[1].Text = s.Name
			r.sessionTableRows[i].Cells[2].Text = s.ModTime.Format("2006-01-02 15:04:05")
		}
		r.sessionTable.SetItems(r.sessionTableRows)
		r.sessionTable.OnItemSelected(func(context *guigui.Context, idx int) {
			if idx < 0 || idx >= len(r.sessions) {
				return
			}
			items, err := types.ParseTimeline(filepath.Join(r.projectDir, r.sessions[idx].ID+".jsonl"))
			if err != nil {
				return
			}
			r.selectedIdx = idx
			r.items = items
			r.currentScreen = screenDetail
		})

	case screenDetail:
		adder.AddWidget(&r.backButton)
		adder.AddWidget(&r.titleText)
		adder.AddWidget(&r.summaryText)
		adder.AddWidget(&r.panel)

		r.backButton.SetText("← Back")
		r.backButton.OnDown(func(context *guigui.Context) {
			r.sessionTable.SelectItemByIndex(-1)
			r.currentScreen = screenList
		})

		title := r.sessions[r.selectedIdx].ID
		if name := r.sessions[r.selectedIdx].Name; name != "" {
			title = name
		}
		r.titleText.SetValue(title)
		r.titleText.SetBold(true)

		var total types.Usage
		for _, item := range r.items {
			if item.Turn != nil {
				u := item.Turn.TotalUsage()
				total.InputTokens += u.InputTokens
				total.OutputTokens += u.OutputTokens
				total.CacheReadInputTokens += u.CacheReadInputTokens
				total.CacheCreationInputTokens += u.CacheCreationInputTokens
			}
		}
		r.summaryText.SetValue(formatUsage(total))
		r.summaryText.SetMultiline(true)

		r.panel.SetContent(&r.msgList)
		r.msgList.items = r.items
	}

	return nil
}

func (r *Root) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	layouter.LayoutWidget(&r.background, widgetBounds.Bounds())

	switch r.currentScreen {
	case screenList:
		r.layoutItems = slices.Delete(r.layoutItems, 0, len(r.layoutItems))
		r.layoutItems = append(r.layoutItems,
			guigui.LinearLayoutItem{Widget: &r.sessionTable, Size: guigui.FlexibleSize(1)},
		)
		(guigui.LinearLayout{
			Direction: guigui.LayoutDirectionVertical,
			Items:     r.layoutItems,
			Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
		}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)

	case screenDetail:
		r.headerItems = slices.Delete(r.headerItems, 0, len(r.headerItems))
		r.headerItems = append(r.headerItems,
			guigui.LinearLayoutItem{Widget: &r.backButton, Size: guigui.FixedSize(u * 3)},
			guigui.LinearLayoutItem{Widget: &r.titleText, Size: guigui.FlexibleSize(1)},
		)
		header := guigui.LinearLayout{
			Direction: guigui.LayoutDirectionHorizontal,
			Items:     r.headerItems,
			Gap:       u / 2,
		}

		r.layoutItems = slices.Delete(r.layoutItems, 0, len(r.layoutItems))
		r.layoutItems = append(r.layoutItems,
			guigui.LinearLayoutItem{Layout: &header, Size: guigui.FixedSize(u)},
			guigui.LinearLayoutItem{Widget: &r.summaryText, Size: guigui.FixedSize(u * 2)},
			guigui.LinearLayoutItem{Widget: &r.panel, Size: guigui.FlexibleSize(1)},
		)
		(guigui.LinearLayout{
			Direction: guigui.LayoutDirectionVertical,
			Items:     r.layoutItems,
			Gap:       u / 2,
			Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
		}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
	}
}

func main() {
	projectDir, err := getProjectDir()
	if err != nil {
		log.Fatal(err)
	}
	sessions, err := loadSessions(projectDir)
	if err != nil {
		log.Fatal(err)
	}

	root := &Root{
		projectDir:  projectDir,
		sessions:    sessions,
		selectedIdx: -1,
	}
	if err := guigui.Run(root, &guigui.RunOptions{
		Title:         "cchist",
		WindowMinSize: image.Pt(960, 600),
		AppScale:      1.5,
	}); err != nil {
		log.Fatal(err)
	}
}
