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

func loadSessions() ([]types.Session, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dirName := strings.ReplaceAll(pwd, "/", "-")
	projectDir := filepath.Join(home, ".claude", "projects", dirName)

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

	background  basicwidget.Background
	table       basicwidget.Table[int]
	tableRows   []basicwidget.TableRow[int]
	layoutItems []guigui.LinearLayoutItem

	sessions []types.Session
}

func (r *Root) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&r.background)
	adder.AddWidget(&r.table)

	r.table.SetItemHeight(basicwidget.UnitSize(context))
	r.table.SetColumns([]basicwidget.TableColumn{
		{
			HeaderText: "id",
			Width:      guigui.FixedSize(basicwidget.UnitSize(context) * 10),
		},
		{
			HeaderText: "name",
			Width:      guigui.FlexibleSize(1),
		},
		{
			HeaderText: "updated",
			Width:      guigui.FixedSize(basicwidget.UnitSize(context) * 6),
		},
	})

	n := len(r.sessions)
	if len(r.tableRows) < n {
		r.tableRows = slices.Grow(r.tableRows, n-len(r.tableRows))[:n]
	} else {
		r.tableRows = slices.Delete(r.tableRows, n, len(r.tableRows))
	}
	for i, s := range r.sessions {
		r.tableRows[i].Value = i
		if len(r.tableRows[i].Cells) < 3 {
			r.tableRows[i].Cells = make([]basicwidget.TableCell, 3)
		}
		r.tableRows[i].Cells[0].Text = s.ID
		r.tableRows[i].Cells[1].Text = s.Name
		r.tableRows[i].Cells[2].Text = s.ModTime.Format("2006-01-02 15:04:05")
	}
	r.table.SetItems(r.tableRows)

	return nil
}

func (r *Root) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	layouter.LayoutWidget(&r.background, widgetBounds.Bounds())

	r.layoutItems = slices.Delete(r.layoutItems, 0, len(r.layoutItems))
	r.layoutItems = append(r.layoutItems,
		guigui.LinearLayoutItem{Widget: &r.table, Size: guigui.FlexibleSize(1)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     r.layoutItems,
		Padding: guigui.Padding{
			Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2,
		},
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func main() {
	sessions, err := loadSessions()
	if err != nil {
		log.Fatal(err)
	}

	root := &Root{sessions: sessions}
	if err := guigui.Run(root, &guigui.RunOptions{
		Title:         "cchist",
		WindowMinSize: image.Pt(960, 600),
		AppScale:      1.5,
	}); err != nil {
		log.Fatal(err)
	}
}
