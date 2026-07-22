package detail

import (
	"slices"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchistory/internal/types"
)

type summaryViewWidget struct {
	guigui.DefaultWidget

	session types.Session

	onBack      func(*guigui.Context)
	onMsgDetail func(*guigui.Context)

	backButton      basicwidget.Button
	titleText       basicwidget.Text
	summaryForm     summaryFormWidget
	msgDetailButton basicwidget.Button

	headerItems []guigui.LinearLayoutItem
	layoutItems []guigui.LinearLayoutItem
}

func (w *summaryViewWidget) setData(session types.Session, items []types.TimelineItem) {
	w.session = session

	var total types.Usage
	for _, item := range items {
		if item.Turn != nil {
			u := item.Turn.TotalUsage()
			total.InputTokens += u.InputTokens
			total.OutputTokens += u.OutputTokens
			total.CacheReadInputTokens += u.CacheReadInputTokens
			total.CacheCreationInputTokens += u.CacheCreationInputTokens
		}
	}
	w.summaryForm.sessionID = session.ID
	w.summaryForm.total = total
}

func (w *summaryViewWidget) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&w.backButton)
	adder.AddWidget(&w.titleText)
	adder.AddWidget(&w.summaryForm)
	adder.AddWidget(&w.msgDetailButton)

	w.backButton.SetText("← Back")
	w.backButton.OnDown(w.onBack)

	title := w.session.ID
	if name := w.session.Name; name != "" {
		title = name
	}
	w.titleText.SetValue(title)
	w.titleText.SetBold(true)

	w.msgDetailButton.SetText("メッセージ詳細")
	w.msgDetailButton.OnDown(w.onMsgDetail)

	return nil
}

func (w *summaryViewWidget) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(ctx)

	w.headerItems = slices.Delete(w.headerItems, 0, len(w.headerItems))
	w.headerItems = append(w.headerItems,
		guigui.LinearLayoutItem{Widget: &w.backButton, Size: guigui.FixedSize(u * 3)},
		guigui.LinearLayoutItem{Widget: &w.titleText, Size: guigui.FlexibleSize(1)},
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
