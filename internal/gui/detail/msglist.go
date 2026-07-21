package detail

import (
	"fmt"
	"image"
	"slices"
	"strings"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchist/internal/types"
)

type turnRowWidget struct {
	guigui.DefaultWidget

	expander    basicwidget.Expander
	headerText  basicwidget.Text
	contentText basicwidget.Text

	turn types.Turn
}

func (w *turnRowWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	firstLine := w.turn.User.Content
	if i := strings.IndexByte(firstLine, '\n'); i >= 0 {
		firstLine = firstLine[:i]
	}
	header := fmt.Sprintf("[%s] %s",
		w.turn.User.Timestamp.Local().Format("2006-01-02 15:04:05.000"),
		firstLine,
	)
	if w.turn.Cancelled() {
		header = "[cancelled] " + header
	}
	w.headerText.SetValue(header)
	w.contentText.SetValue(w.buildContent())
	w.contentText.SetWrapMode(basicwidget.WrapModeNormal)
	w.contentText.SetMultiline(true)
	w.contentText.SetSelectable(true)
	w.expander.SetHeaderWidget(&w.headerText)
	w.expander.SetContentWidget(&w.contentText)
	adder.AddWidget(&w.expander)
	return nil
}

func (w *turnRowWidget) buildContent() string {
	var sb strings.Builder

	sb.WriteString("User:\n")
	sb.WriteString(w.turn.User.Content)

	if w.turn.Cancelled() {
		return strings.TrimRight(sb.String(), "\n")
	}

	sb.WriteString("\n\n──────────────────────────\n\n")

	sb.WriteString("Assistant:\n")
	if content := w.turn.AssistantContent(); content != "" {
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("──────────────────────────\n")

	if t := w.turn.FinishedAt(); !t.IsZero() {
		sb.WriteString(fmt.Sprintf("Finished: %s\n", t.Local().Format("2006-01-02 15:04:05.000")))
	}

	var meta []string
	if m := w.turn.Model(); m != "" {
		meta = append(meta, "Model: "+m)
	}
	if e := w.turn.Effort(); e != "" {
		meta = append(meta, "Effort: "+e)
	}
	if len(meta) > 0 {
		sb.WriteString(strings.Join(meta, "  "))
		sb.WriteString("\n")
	}

	sb.WriteString(formatUsage(w.turn.TotalUsage()))

	return strings.TrimRight(sb.String(), "\n")
}

func (w *turnRowWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&w.expander, widgetBounds.Bounds())
}

func (w *turnRowWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.expander.Measure(context, constraints)
}

type compactRowWidget struct {
	guigui.DefaultWidget
	expander    basicwidget.Expander
	headerText  basicwidget.Text
	contentText basicwidget.Text
	cb          types.CompactBoundary
}

func (w *compactRowWidget) buildContent() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"%s → %s tokens  dropped: %s  took: %.1fs",
		formatTokens(w.cb.PreTokens),
		formatTokens(w.cb.PostTokens),
		formatTokens(w.cb.DroppedTokens),
		float64(w.cb.DurationMs)/1000,
	))
	if w.cb.Summary != "" {
		sb.WriteString("\n\n──────────────────────────\n\n")
		sb.WriteString(w.cb.Summary)
	}
	return sb.String()
}

func (w *compactRowWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	w.headerText.SetValue(fmt.Sprintf("[Compacted: %s]  %s",
		w.cb.Trigger,
		w.cb.Timestamp.Local().Format("2006-01-02 15:04:05"),
	))
	w.contentText.SetValue(w.buildContent())
	w.contentText.SetWrapMode(basicwidget.WrapModeNormal)
	w.contentText.SetMultiline(true)
	w.contentText.SetSelectable(true)
	w.expander.SetHeaderWidget(&w.headerText)
	w.expander.SetContentWidget(&w.contentText)
	adder.AddWidget(&w.expander)
	return nil
}

func (w *compactRowWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&w.expander, widgetBounds.Bounds())
}

func (w *compactRowWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.expander.Measure(context, constraints)
}

type localCmdRowWidget struct {
	guigui.DefaultWidget
	expander    basicwidget.Expander
	headerText  basicwidget.Text
	contentText basicwidget.Text
	cmd         types.LocalCommand
}

func (w *localCmdRowWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	w.headerText.SetValue(fmt.Sprintf("[%s] $ %s",
		w.cmd.Timestamp.Local().Format("2006-01-02 15:04:05.000"),
		w.cmd.Input,
	))

	var out strings.Builder
	if w.cmd.Stdout != "" {
		out.WriteString(w.cmd.Stdout)
	}
	if w.cmd.Stderr != "" {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		out.WriteString(w.cmd.Stderr)
	}

	if out.Len() > 0 {
		w.contentText.SetValue(out.String())
		w.contentText.SetWrapMode(basicwidget.WrapModeNormal)
		w.contentText.SetMultiline(true)
		w.contentText.SetSelectable(true)
		w.expander.SetHeaderWidget(&w.headerText)
		w.expander.SetContentWidget(&w.contentText)
		adder.AddWidget(&w.expander)
	} else {
		adder.AddWidget(&w.headerText)
	}
	return nil
}

func (w *localCmdRowWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	if w.cmd.Stdout != "" || w.cmd.Stderr != "" {
		layouter.LayoutWidget(&w.expander, widgetBounds.Bounds())
	} else {
		layouter.LayoutWidget(&w.headerText, widgetBounds.Bounds())
	}
}

func (w *localCmdRowWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	if w.cmd.Stdout != "" || w.cmd.Stderr != "" {
		return w.expander.Measure(context, constraints)
	}
	return w.headerText.Measure(context, constraints)
}

type msgListWidget struct {
	guigui.DefaultWidget

	rows         guigui.WidgetSlice[*turnRowWidget]
	compactRows  guigui.WidgetSlice[*compactRowWidget]
	localCmdRows guigui.WidgetSlice[*localCmdRowWidget]
	dividers     guigui.WidgetSlice[*basicwidget.Divider]
	layoutItems  []guigui.LinearLayoutItem

	items []types.TimelineItem
}

func (w *msgListWidget) layout(context *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(context)
	n := len(w.items)
	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	ti, ci, li := 0, 0, 0
	for i, item := range w.items {
		var widget guigui.Widget
		switch {
		case item.Turn != nil:
			widget = w.rows.At(ti)
			ti++
		case item.CompactBoundary != nil:
			widget = w.compactRows.At(ci)
			ci++
		default:
			widget = w.localCmdRows.At(li)
			li++
		}
		w.layoutItems = append(w.layoutItems, guigui.LinearLayoutItem{Widget: widget})
		if i < n-1 {
			w.layoutItems = append(w.layoutItems, guigui.LinearLayoutItem{Widget: w.dividers.At(i), Size: guigui.FixedSize(1)})
		}
	}
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
	}
}

func (w *msgListWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	n := len(w.items)

	turnCount, compactCount, localCmdCount := 0, 0, 0
	for _, item := range w.items {
		switch {
		case item.Turn != nil:
			turnCount++
		case item.CompactBoundary != nil:
			compactCount++
		default:
			localCmdCount++
		}
	}
	w.rows.SetLen(turnCount)
	w.compactRows.SetLen(compactCount)
	w.localCmdRows.SetLen(localCmdCount)
	w.dividers.SetLen(max(0, n-1))

	ti, ci, li := 0, 0, 0
	for i, item := range w.items {
		switch {
		case item.Turn != nil:
			w.rows.At(ti).turn = *item.Turn
			adder.AddWidget(w.rows.At(ti))
			ti++
		case item.CompactBoundary != nil:
			w.compactRows.At(ci).cb = *item.CompactBoundary
			adder.AddWidget(w.compactRows.At(ci))
			ci++
		default:
			w.localCmdRows.At(li).cmd = *item.LocalCommand
			adder.AddWidget(w.localCmdRows.At(li))
			li++
		}
		if i < n-1 {
			adder.AddWidget(w.dividers.At(i))
		}
	}
	return nil
}

func (w *msgListWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	w.layout(context).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (w *msgListWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.layout(context).Measure(context, constraints)
}
