package gui

import (
	"path/filepath"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchistory/internal/gui/detail"
	"github.com/masakurapa/cchistory/internal/gui/list"
	"github.com/masakurapa/cchistory/internal/types"
)

type root struct {
	guigui.DefaultWidget
	background basicwidget.Background

	list    *list.Widget
	detail  *detail.Widget
	current Widget
}

func newRoot(projectDir string, sessions []types.Session) *root {
	r := &root{
		list:   list.New(),
		detail: detail.New(),
	}
	r.current = r.list
	r.initListWidget(projectDir, sessions)
	r.initDetailWidget()
	return r
}

func (r *root) initListWidget(projectDir string, sessions []types.Session) {
	r.list.SetSessions(sessions)
	r.list.SetOnSelected(func(ctx *guigui.Context, session types.Session) {
		items, err := types.ParseTimeline(filepath.Join(projectDir, session.ID+".jsonl"))
		if err != nil {
			return
		}
		r.detail.SetData(session, items)
		r.current = r.detail
	})
}

func (r *root) initDetailWidget() {
	r.detail.SetOnBack(func(ctx *guigui.Context) {
		r.list.ResetSelection()
		r.current = r.list
	})
}

func (r *root) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&r.background)
	return r.current.Build(context, adder)
}

func (r *root) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&r.background, widgetBounds.Bounds())
	r.current.Layout(context, widgetBounds, layouter)
}
