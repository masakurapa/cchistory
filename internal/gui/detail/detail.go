package detail

import (
	"github.com/guigui-gui/guigui"
	"github.com/masakurapa/cchistory/internal/types"
)

type view interface {
	Build(*guigui.Context, *guigui.ChildAdder) error
	Layout(*guigui.Context, *guigui.WidgetBounds, *guigui.ChildLayouter)
}

type Widget struct {
	guigui.DefaultWidget

	summaryView summaryViewWidget
	msgDetail   msgDetailWidget
	current     view

	onBack func(*guigui.Context)
}

func New() *Widget {
	w := &Widget{}
	w.current = &w.summaryView
	return w
}

func (w *Widget) SetData(session types.Session, items []types.TimelineItem) {
	w.summaryView.setData(session, items)
	w.summaryView.onBack = w.onBack
	w.summaryView.onMsgDetail = func(ctx *guigui.Context) {
		w.msgDetail.items = items
		w.msgDetail.selectedItemIdx = -1
		w.msgDetail.onBack = func(ctx *guigui.Context) {
			w.current = &w.summaryView
		}
		w.current = &w.msgDetail
	}
	w.current = &w.summaryView
}

func (w *Widget) SetOnBack(fn func(*guigui.Context)) {
	w.onBack = fn
	w.summaryView.onBack = fn
}

func (w *Widget) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	return w.current.Build(ctx, adder)
}

func (w *Widget) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	w.current.Layout(ctx, wb, layouter)
}
