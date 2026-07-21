package detail

import (
	"slices"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchist/internal/types"
)

type Widget struct {
	guigui.DefaultWidget

	session types.Session
	items   []types.TimelineItem
	onBack  func(*guigui.Context)

	backButton  basicwidget.Button
	titleText   basicwidget.Text
	summaryText basicwidget.Text
	panel       basicwidget.Panel
	msgList     msgListWidget
	layoutItems []guigui.LinearLayoutItem
	headerItems []guigui.LinearLayoutItem
}

func New() *Widget {
	return &Widget{}
}

func (w *Widget) SetData(session types.Session, items []types.TimelineItem) {
	w.session = session
	w.items = items
}

func (w *Widget) SetOnBack(fn func(*guigui.Context)) {
	w.onBack = fn
}

func (w *Widget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&w.backButton)
	adder.AddWidget(&w.titleText)
	adder.AddWidget(&w.summaryText)
	adder.AddWidget(&w.panel)

	w.backButton.SetText("← Back")
	w.backButton.OnDown(w.onBack)

	title := w.session.ID
	if name := w.session.Name; name != "" {
		title = name
	}
	w.titleText.SetValue(title)
	w.titleText.SetBold(true)

	var total types.Usage
	for _, item := range w.items {
		if item.Turn != nil {
			u := item.Turn.TotalUsage()
			total.InputTokens += u.InputTokens
			total.OutputTokens += u.OutputTokens
			total.CacheReadInputTokens += u.CacheReadInputTokens
			total.CacheCreationInputTokens += u.CacheCreationInputTokens
		}
	}
	w.summaryText.SetValue(formatUsage(total))
	w.summaryText.SetMultiline(true)

	w.msgList.items = w.items
	w.panel.SetContent(&w.msgList)

	return nil
}

func (w *Widget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)

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
		guigui.LinearLayoutItem{Widget: &w.summaryText, Size: guigui.FixedSize(u * 2)},
		guigui.LinearLayoutItem{Widget: &w.panel, Size: guigui.FlexibleSize(1)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Gap:       u / 2,
		Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}
