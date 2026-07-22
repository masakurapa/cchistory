package detail

import (
	"slices"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchistory/internal/types"
)

type summaryViewWidget struct {
	guigui.DefaultWidget

	onBack      func(*guigui.Context)
	onMsgDetail func(*guigui.Context)

	backButton      basicwidget.Button
	screenLabel     basicwidget.Text
	summaryForm     metaFormWidget
	msgDetailButton basicwidget.Button

	headerItems []guigui.LinearLayoutItem
	layoutItems []guigui.LinearLayoutItem
}

func (w *summaryViewWidget) setData(d types.SessionDetail) {
	w.summaryForm.metas = d.Metas
}

func (w *summaryViewWidget) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&w.backButton)
	adder.AddWidget(&w.screenLabel)
	adder.AddWidget(&w.summaryForm)
	adder.AddWidget(&w.msgDetailButton)

	w.backButton.SetText("← Back")
	w.backButton.OnDown(w.onBack)

	w.screenLabel.SetValue("Session Summary")
	w.screenLabel.SetBold(true)

	w.msgDetailButton.SetText("Messages")
	w.msgDetailButton.OnDown(w.onMsgDetail)

	return nil
}

func (w *summaryViewWidget) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(ctx)

	w.headerItems = slices.Delete(w.headerItems, 0, len(w.headerItems))
	w.headerItems = append(w.headerItems,
		guigui.LinearLayoutItem{Widget: &w.backButton, Size: guigui.FixedSize(u * 3)},
		guigui.LinearLayoutItem{Widget: &w.screenLabel, Size: guigui.FlexibleSize(1)},
	)
	header := guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     w.headerItems,
		Gap:       u / 2,
	}

	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	w.layoutItems = append(w.layoutItems,
		guigui.LinearLayoutItem{Layout: &header, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Widget: &w.summaryForm},
		guigui.LinearLayoutItem{Widget: &w.msgDetailButton, Size: guigui.FixedSize(u * 2)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Gap:       u / 2,
		Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
	}).LayoutWidgets(ctx, wb.Bounds(), layouter)
}
