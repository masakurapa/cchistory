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

// metaFormWidget displays a []Meta slice as a macOS-settings-style Form.
type metaFormWidget struct {
	guigui.DefaultWidget

	form      basicwidget.Form
	metas     []types.Meta
	labels    guigui.WidgetSlice[*basicwidget.Text]
	values    guigui.WidgetSlice[*basicwidget.Text]
	formItems []basicwidget.FormItem
}

func (w *metaFormWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	w.labels.SetLen(len(w.metas))
	w.values.SetLen(len(w.metas))
	w.formItems = w.formItems[:0]

	for i, meta := range w.metas {
		lbl := w.labels.At(i)
		val := w.values.At(i)
		lbl.SetValue(meta.Name)
		val.SetValue(meta.Value)
		val.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
		val.SetSelectable(true)
		w.formItems = append(w.formItems, basicwidget.FormItem{
			PrimaryWidget:   lbl,
			SecondaryWidget: val,
		})
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

