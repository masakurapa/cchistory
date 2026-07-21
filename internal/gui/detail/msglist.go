package detail

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchistory/internal/types"
)

const (
	textAreaMinLines = 3
	textAreaMaxLines = 5
)

// textAreaContent is the panel content: a Text widget with horizontal padding.
type textAreaContent struct {
	guigui.DefaultWidget
	text basicwidget.Text
}

func (w *textAreaContent) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&w.text)
	return nil
}

func (w *textAreaContent) hPad(context *guigui.Context) int {
	return basicwidget.UnitSize(context) / 2
}

func (w *textAreaContent) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	p := w.hPad(context)
	b := widgetBounds.Bounds()
	layouter.LayoutWidget(&w.text, image.Rect(b.Min.X+p, b.Min.Y, b.Max.X-p, b.Max.Y))
}

func (w *textAreaContent) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	p := w.hPad(context)
	inner := constraints
	if fw, ok := constraints.FixedWidth(); ok {
		inner = guigui.FixedWidthConstraints(max(0, fw-p*2))
	}
	m := w.text.Measure(context, inner)
	return image.Pt(m.X+p*2, m.Y)
}

// textAreaWidget is a scrollable read-only text box with a white background.
type textAreaWidget struct {
	guigui.DefaultWidget

	panel   basicwidget.Panel
	content textAreaContent
}

func (w *textAreaWidget) setText(s string) {
	w.content.text.SetValue(s)
	w.content.text.SetWrapMode(basicwidget.WrapModeNormal)
	w.content.text.SetMultiline(true)
	w.content.text.SetSelectable(true)
}

func (w *textAreaWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	w.panel.SetContent(&w.content)
	w.panel.SetContentConstraints(basicwidget.PanelContentConstraintsFixedWidth)
	w.panel.SetBackgroundStyle(basicwidget.PanelBackgroundStyleNone)
	w.panel.SetBorders(basicwidget.PanelBorders{Start: true, Top: true, End: true, Bottom: true})
	adder.AddWidget(&w.panel)
	return nil
}

func (w *textAreaWidget) Draw(_ *guigui.Context, widgetBounds *guigui.WidgetBounds, dst *ebiten.Image) {
	dst.SubImage(widgetBounds.Bounds()).(*ebiten.Image).Fill(color.White)
}

func (w *textAreaWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&w.panel, widgetBounds.Bounds())
}

func (w *textAreaWidget) height(context *guigui.Context, width int) int {
	u := basicwidget.UnitSize(context)
	minH := u * textAreaMinLines
	maxH := u * textAreaMaxLines
	measured := w.content.Measure(context, guigui.FixedWidthConstraints(width))
	return max(minH, min(measured.Y, maxH))
}

func (w *textAreaWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	width, ok := constraints.FixedWidth()
	if !ok {
		width = 800
	}
	return image.Pt(width, w.height(context, width))
}

// metaFormWidget displays turn metadata in a macOS-settings-style Form.
type metaFormWidget struct {
	guigui.DefaultWidget

	form basicwidget.Form
	turn types.Turn

	finishedLabel basicwidget.Text
	finishedValue basicwidget.Text
	modelLabel    basicwidget.Text
	modelValue    basicwidget.Text
	effortLabel   basicwidget.Text
	effortValue   basicwidget.Text
	inputLabel    basicwidget.Text
	inputValue    basicwidget.Text
	outputLabel   basicwidget.Text
	outputValue   basicwidget.Text
	cacheRLabel   basicwidget.Text
	cacheRValue   basicwidget.Text
	cacheCLabel   basicwidget.Text
	cacheCValue   basicwidget.Text

	formItems []basicwidget.FormItem
}

func (w *metaFormWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	usage := w.turn.TotalUsage()
	w.formItems = w.formItems[:0]

	row := func(label *basicwidget.Text, labelText string, value *basicwidget.Text, valueText string) {
		label.SetValue(labelText)
		value.SetValue(valueText)
		value.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
		w.formItems = append(w.formItems, basicwidget.FormItem{
			PrimaryWidget:   label,
			SecondaryWidget: value,
		})
	}

	if t := w.turn.FinishedAt(); !t.IsZero() {
		row(&w.finishedLabel, "Finished", &w.finishedValue, t.Local().Format("2006-01-02 15:04:05"))
	}
	if m := w.turn.Model(); m != "" {
		row(&w.modelLabel, "Model", &w.modelValue, m)
	}
	if e := w.turn.Effort(); e != "" {
		row(&w.effortLabel, "Effort", &w.effortValue, e)
	}
	row(&w.inputLabel, "Input Token", &w.inputValue, formatTokens(usage.InputTokens))
	row(&w.outputLabel, "Output Token", &w.outputValue, formatTokens(usage.OutputTokens))
	if usage.CacheReadInputTokens > 0 {
		row(&w.cacheRLabel, "Cache Read", &w.cacheRValue, formatTokens(usage.CacheReadInputTokens))
	}
	if usage.CacheCreationInputTokens > 0 {
		row(&w.cacheCLabel, "Cache Creation", &w.cacheCValue, formatTokens(usage.CacheCreationInputTokens))
	}

	w.form.SetItems(w.formItems)
	adder.AddWidget(&w.form)
	return nil
}

func (w *metaFormWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&w.form, widgetBounds.Bounds())
}

func (w *metaFormWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.form.Measure(context, constraints)
}

// turnContentWidget is the content area inside the Expander for a single turn.
type turnContentWidget struct {
	guigui.DefaultWidget

	userLabel   basicwidget.Text
	userArea    textAreaWidget
	assistLabel basicwidget.Text
	assistArea  textAreaWidget
	metaForm    metaFormWidget
	layoutItems []guigui.LinearLayoutItem

	turn types.Turn
}

func (w *turnContentWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	w.userLabel.SetValue("User:")
	adder.AddWidget(&w.userLabel)
	adder.AddWidget(&w.userArea)

	if !w.turn.Cancelled() {
		w.assistLabel.SetValue("Assistant:")
		w.metaForm.turn = w.turn
		adder.AddWidget(&w.assistLabel)
		adder.AddWidget(&w.assistArea)
		adder.AddWidget(&w.metaForm)
	}
	return nil
}

func (w *turnContentWidget) buildLayout(context *guigui.Context, width int) guigui.LinearLayout {
	u := basicwidget.UnitSize(context)
	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	w.layoutItems = append(w.layoutItems,
		guigui.LinearLayoutItem{Widget: &w.userLabel, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Widget: &w.userArea, Size: guigui.FixedSize(w.userArea.height(context, width))},
	)
	if !w.turn.Cancelled() {
		w.layoutItems = append(w.layoutItems,
			guigui.LinearLayoutItem{Widget: &w.assistLabel, Size: guigui.FixedSize(u)},
			guigui.LinearLayoutItem{Widget: &w.assistArea, Size: guigui.FixedSize(w.assistArea.height(context, width))},
			guigui.LinearLayoutItem{Widget: &w.metaForm},
		)
	}
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Gap:       u / 4,
		Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
	}
}

func (w *turnContentWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	b := widgetBounds.Bounds()
	w.buildLayout(context, b.Dx()).LayoutWidgets(context, b, layouter)
}

func (w *turnContentWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	width, ok := constraints.FixedWidth()
	if !ok {
		width = 800
	}
	return w.buildLayout(context, width).Measure(context, constraints)
}

type turnRowWidget struct {
	guigui.DefaultWidget

	expander   basicwidget.Expander
	headerText basicwidget.Text
	content    turnContentWidget

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

	w.content.turn = w.turn
	w.content.userArea.setText(w.turn.User.Content)
	if !w.turn.Cancelled() {
		w.content.assistArea.setText(w.turn.AssistantContent())
	}

	w.expander.SetHeaderWidget(&w.headerText)
	w.expander.SetContentWidget(&w.content)
	adder.AddWidget(&w.expander)
	return nil
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
