package detail

import (
	"image"
	"image/color"

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
		value.SetSelectable(true)
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

// summaryFormWidget displays session total usage in macOS-settings-style Form.
type summaryFormWidget struct {
	guigui.DefaultWidget

	form      basicwidget.Form
	total     types.Usage
	sessionID string

	idLabel     basicwidget.Text
	idValue     basicwidget.Text
	inputLabel  basicwidget.Text
	inputValue  basicwidget.Text
	outputLabel basicwidget.Text
	outputValue basicwidget.Text
	cacheRLabel basicwidget.Text
	cacheRValue basicwidget.Text
	cacheCLabel basicwidget.Text
	cacheCValue basicwidget.Text

	formItems []basicwidget.FormItem
}

func (w *summaryFormWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	w.formItems = w.formItems[:0]

	row := func(label *basicwidget.Text, labelText string, value *basicwidget.Text, valueText string) {
		label.SetValue(labelText)
		value.SetValue(valueText)
		value.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
		value.SetSelectable(true)
		w.formItems = append(w.formItems, basicwidget.FormItem{
			PrimaryWidget:   label,
			SecondaryWidget: value,
		})
	}

	row(&w.idLabel, "ID", &w.idValue, w.sessionID)
	row(&w.inputLabel, "Input Token", &w.inputValue, formatTokens(w.total.InputTokens))
	row(&w.outputLabel, "Output Token", &w.outputValue, formatTokens(w.total.OutputTokens))
	if w.total.CacheReadInputTokens > 0 {
		row(&w.cacheRLabel, "Cache Read", &w.cacheRValue, formatTokens(w.total.CacheReadInputTokens))
	}
	if w.total.CacheCreationInputTokens > 0 {
		row(&w.cacheCLabel, "Cache Creation", &w.cacheCValue, formatTokens(w.total.CacheCreationInputTokens))
	}

	w.form.SetItems(w.formItems)
	adder.AddWidget(&w.form)
	return nil
}

func (w *summaryFormWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&w.form, widgetBounds.Bounds())
}

func (w *summaryFormWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.form.Measure(context, constraints)
}
