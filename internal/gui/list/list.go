package list

import (
	"slices"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchist/internal/types"
)

type Widget struct {
	sessions   []types.Session
	onSelected func(*guigui.Context, types.Session)

	sessionTable     basicwidget.Table[int]
	sessionTableRows []basicwidget.TableRow[int]
	layoutItems      []guigui.LinearLayoutItem
}

func New() *Widget {
	return &Widget{}
}

func (w *Widget) SetSessions(sessions []types.Session) {
	w.sessions = sessions
}

func (w *Widget) SetOnSelected(fn func(*guigui.Context, types.Session)) {
	w.onSelected = fn
}

func (w *Widget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	u := basicwidget.UnitSize(context)
	adder.AddWidget(&w.sessionTable)

	w.sessionTable.SetItemHeight(u)
	w.sessionTable.SetColumns([]basicwidget.TableColumn{
		{HeaderText: "id", Width: guigui.FixedSize(u * 10)},
		{HeaderText: "name", Width: guigui.FlexibleSize(1)},
		{HeaderText: "updated", Width: guigui.FixedSize(u * 6)},
	})

	n := len(w.sessions)
	if len(w.sessionTableRows) < n {
		w.sessionTableRows = slices.Grow(w.sessionTableRows, n-len(w.sessionTableRows))[:n]
	} else {
		w.sessionTableRows = slices.Delete(w.sessionTableRows, n, len(w.sessionTableRows))
	}
	for i, s := range w.sessions {
		w.sessionTableRows[i].Value = i
		if len(w.sessionTableRows[i].Cells) < 3 {
			w.sessionTableRows[i].Cells = make([]basicwidget.TableCell, 3)
		}
		w.sessionTableRows[i].Cells[0].Text = s.ID
		w.sessionTableRows[i].Cells[1].Text = s.Name
		w.sessionTableRows[i].Cells[2].Text = s.ModTime.Format("2006-01-02 15:04:05")
	}

	w.sessionTable.SetItems(w.sessionTableRows)
	w.sessionTable.OnItemSelected(func(context *guigui.Context, idx int) {
		if idx < 0 || idx >= len(w.sessions) {
			return
		}
		w.onSelected(context, w.sessions[idx])
	})

	return nil
}

func (w *Widget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)

	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	w.layoutItems = append(w.layoutItems,
		guigui.LinearLayoutItem{Widget: &w.sessionTable, Size: guigui.FlexibleSize(1)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}
