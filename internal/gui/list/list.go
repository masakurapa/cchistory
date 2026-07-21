package list

import (
	"cmp"
	"slices"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchistory/internal/types"
)

type Widget struct {
	sessions   []types.Session
	onSelected func(*guigui.Context, types.Session)

	headerName  basicwidget.Text
	headerDate  basicwidget.Text
	headerItems []guigui.LinearLayoutItem

	sessionList     basicwidget.List[int]
	sessionListRows []basicwidget.ListItem[int]
	layoutItems     []guigui.LinearLayoutItem
}

func New() *Widget {
	return &Widget{}
}

func (w *Widget) SetSessions(sessions []types.Session) {
	w.sessions = slices.SortedFunc(slices.Values(sessions), func(a, b types.Session) int {
		return cmp.Compare(b.ModTime.UnixNano(), a.ModTime.UnixNano())
	})
}

func (w *Widget) SetOnSelected(fn func(*guigui.Context, types.Session)) {
	w.onSelected = fn
}

func (w *Widget) ResetSelection() {
	w.sessionList.SelectItemByIndex(-1)
}

func (w *Widget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	w.headerName.SetValue("Name")
	w.headerName.SetBold(true)
	w.headerDate.SetValue("Updated")
	w.headerDate.SetBold(true)
	w.headerDate.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	adder.AddWidget(&w.headerName)
	adder.AddWidget(&w.headerDate)

	adder.AddWidget(&w.sessionList)
	w.sessionList.SetStyle(basicwidget.ListStyleNormal)

	n := len(w.sessions)
	total := n + max(0, n-1)
	w.sessionListRows = slices.Delete(w.sessionListRows, 0, len(w.sessionListRows))
	w.sessionListRows = slices.Grow(w.sessionListRows, total)

	for i, s := range w.sessions {
		label := s.Name
		if label == "" {
			label = s.ID
		}
		w.sessionListRows = append(w.sessionListRows, basicwidget.ListItem[int]{
			Text:    label,
			KeyText: s.ModTime.Format("2006-01-02 15:04"),
			Value:   i,
		})
		if i < n-1 {
			w.sessionListRows = append(w.sessionListRows, basicwidget.ListItem[int]{Border: true})
		}
	}

	w.sessionList.SetItems(w.sessionListRows)
	w.sessionList.OnItemSelected(func(context *guigui.Context, idx int) {
		item, ok := w.sessionList.ItemByIndex(idx)
		if !ok {
			return
		}
		sessionIdx := item.Value
		if sessionIdx < 0 || sessionIdx >= len(w.sessions) {
			return
		}
		w.onSelected(context, w.sessions[sessionIdx])
	})

	return nil
}

func (w *Widget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	// u/2 (outer padding) + u/4 (rounded corner) + u/4 (item text padding) = u
	// The outer layout already adds u/2, so add u/2 more to align with list item text.
	hPad := u / 2

	w.headerItems = slices.Delete(w.headerItems, 0, len(w.headerItems))
	w.headerItems = append(w.headerItems,
		guigui.LinearLayoutItem{Widget: &w.headerName, Size: guigui.FlexibleSize(1)},
		guigui.LinearLayoutItem{Widget: &w.headerDate},
	)
	header := guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     w.headerItems,
		Padding:   guigui.Padding{Start: hPad, End: hPad},
	}

	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	w.layoutItems = append(w.layoutItems,
		guigui.LinearLayoutItem{Layout: &header, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Widget: &w.sessionList, Size: guigui.FlexibleSize(1)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Gap:       u / 4,
		Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}
