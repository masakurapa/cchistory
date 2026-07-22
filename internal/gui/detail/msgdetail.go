package detail

import (
	"image"
	"image/color"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchistory/internal/types"
)

// coloredDivider draws a vertical line whose widget width provides left/right margin.
type coloredDivider struct {
	guigui.DefaultWidget
}

func (d *coloredDivider) Draw(ctx *guigui.Context, wb *guigui.WidgetBounds, dst *ebiten.Image) {
	b := wb.Bounds()
	strokeWidth := float32(1 * ctx.Scale())
	x := float32(b.Min.X+b.Max.X) / 2
	var clr color.RGBA
	if ctx.ColorMode() == ebiten.ColorModeDark {
		clr = color.RGBA{R: 51, G: 51, B: 51, A: 255}
	} else {
		clr = color.RGBA{R: 204, G: 204, B: 204, A: 255}
	}
	vector.StrokeLine(dst, x, float32(b.Min.Y), x, float32(b.Max.Y), strokeWidth, clr, false)
}

// entryListRow is a 2-line clickable row for any TimelineEntry.
type entryListRow struct {
	guigui.DefaultWidget
	entry          types.TimelineEntry
	itemIdx        int
	selectedIdxPtr *int
	onSelected     func(itemIdx int)

	dateText basicwidget.Text
	msgText  basicwidget.Text
}

func (w *entryListRow) WriteStateKey(_ *guigui.Context, sw *guigui.StateKeyWriter) {
	sw.WriteBool(w.selectedIdxPtr != nil && *w.selectedIdxPtr == w.itemIdx)
}

func (w *entryListRow) Build(_ *guigui.Context, adder *guigui.ChildAdder) error {
	w.dateText.SetValue(w.entry.Timestamp().Local().Format("2006-01-02 15:04"))
	w.msgText.SetValue(w.entry.Headline())
	adder.AddWidget(&w.dateText)
	adder.AddWidget(&w.msgText)
	return nil
}

func (w *entryListRow) Draw(_ *guigui.Context, wb *guigui.WidgetBounds, dst *ebiten.Image) {
	if w.selectedIdxPtr != nil && *w.selectedIdxPtr == w.itemIdx {
		dst.SubImage(wb.Bounds()).(*ebiten.Image).Fill(color.RGBA{R: 180, G: 180, B: 180, A: 80})
	}
}

func (w *entryListRow) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(ctx)
	b := wb.Bounds()
	hPad := u / 4
	mid := b.Min.Y + u
	layouter.LayoutWidget(&w.dateText, image.Rect(b.Min.X+hPad, b.Min.Y, b.Max.X-hPad, mid))
	layouter.LayoutWidget(&w.msgText, image.Rect(b.Min.X+hPad, mid, b.Max.X-hPad, b.Max.Y))
}

func (w *entryListRow) Measure(ctx *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(ctx)
	width, ok := constraints.FixedWidth()
	if !ok {
		width = 400
	}
	return image.Pt(width, u*2)
}

func (w *entryListRow) HandlePointingInput(_ *guigui.Context, wb *guigui.WidgetBounds) guigui.HandleInputResult {
	if wb.IsHitAtCursor() && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if w.onSelected != nil {
			w.onSelected(w.itemIdx)
		}
		return guigui.HandleInputByWidget(w)
	}
	return guigui.HandleInputResult{}
}

// turnListContent is the scrollable content of the left panel.
type turnListContent struct {
	guigui.DefaultWidget
	items       []types.TimelineEntry
	selectedIdx int
	onSelected  func(int)

	rows        guigui.WidgetSlice[*entryListRow]
	dividers    guigui.WidgetSlice[*basicwidget.Divider]
	layoutItems []guigui.LinearLayoutItem
}

func (w *turnListContent) Build(_ *guigui.Context, adder *guigui.ChildAdder) error {
	n := len(w.items)
	w.rows.SetLen(n)
	w.dividers.SetLen(max(0, n-1))

	for i, entry := range w.items {
		if i > 0 {
			adder.AddWidget(w.dividers.At(i - 1))
		}
		row := w.rows.At(i)
		row.entry = entry
		row.itemIdx = i
		row.selectedIdxPtr = &w.selectedIdx
		row.onSelected = w.onSelected
		adder.AddWidget(row)
	}
	return nil
}

func (w *turnListContent) layout(ctx *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(ctx)
	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	for i := range w.items {
		if i > 0 {
			w.layoutItems = append(w.layoutItems,
				guigui.LinearLayoutItem{Widget: w.dividers.At(i - 1), Size: guigui.FixedSize(1)},
			)
		}
		w.layoutItems = append(w.layoutItems,
			guigui.LinearLayoutItem{Widget: w.rows.At(i)},
		)
	}
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Padding:   guigui.Padding{Top: u / 2, Bottom: u / 2},
	}
}

func (w *turnListContent) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	w.layout(ctx).LayoutWidgets(ctx, wb.Bounds(), layouter)
}

func (w *turnListContent) Measure(ctx *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.layout(ctx).Measure(ctx, constraints)
}

// sectionWidget renders a label header and a scrollable text area for one Section.
type sectionWidget struct {
	guigui.DefaultWidget
	label       basicwidget.Text
	area        textAreaWidget
	layoutItems []guigui.LinearLayoutItem
}

func (w *sectionWidget) set(labelText, text string) {
	w.label.SetValue(labelText + ":")
	w.area.setText(text)
}

func (w *sectionWidget) Build(_ *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&w.label)
	adder.AddWidget(&w.area)
	return nil
}

func (w *sectionWidget) layout(ctx *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(ctx)
	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	w.layoutItems = append(w.layoutItems,
		guigui.LinearLayoutItem{Widget: &w.label, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Widget: &w.area},
	)
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Gap:       u / 4,
	}
}

func (w *sectionWidget) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	w.layout(ctx).LayoutWidgets(ctx, wb.Bounds(), layouter)
}

func (w *sectionWidget) Measure(ctx *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.layout(ctx).Measure(ctx, constraints)
}

// entryDetailWidget renders the full detail of any TimelineEntry.
type entryDetailWidget struct {
	guigui.DefaultWidget
	entry    types.TimelineEntry

	sections    guigui.WidgetSlice[*sectionWidget]
	metaForm    metaFormWidget
	layoutItems []guigui.LinearLayoutItem
}

func (w *entryDetailWidget) Build(_ *guigui.Context, adder *guigui.ChildAdder) error {
	secs := w.entry.Sections()
	w.sections.SetLen(len(secs))
	for i, s := range secs {
		sw := w.sections.At(i)
		sw.set(s.Label, s.Text)
		adder.AddWidget(sw)
	}
	w.metaForm.metas = w.entry.Metadata()
	adder.AddWidget(&w.metaForm)
	return nil
}

func (w *entryDetailWidget) layout(ctx *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(ctx)
	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	secs := w.entry.Sections()
	for i := range secs {
		w.layoutItems = append(w.layoutItems,
			guigui.LinearLayoutItem{Widget: w.sections.At(i)},
		)
	}
	w.layoutItems = append(w.layoutItems,
		guigui.LinearLayoutItem{Widget: &w.metaForm},
	)
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Gap:       u / 4,
		Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
	}
}

func (w *entryDetailWidget) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	w.layout(ctx).LayoutWidgets(ctx, wb.Bounds(), layouter)
}

func (w *entryDetailWidget) Measure(ctx *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.layout(ctx).Measure(ctx, constraints)
}

type msgDetailWidget struct {
	guigui.DefaultWidget

	items           []types.TimelineEntry
	selectedItemIdx int
	onBack          func(*guigui.Context)

	backButton  basicwidget.Button
	leftPanel   basicwidget.Panel
	listContent turnListContent
	rightPanel  basicwidget.Panel
	rightContent entryDetailWidget
	divider     coloredDivider

	headerItems []guigui.LinearLayoutItem
	bodyItems   []guigui.LinearLayoutItem
	layoutItems []guigui.LinearLayoutItem
}

func (w *msgDetailWidget) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&w.backButton)
	adder.AddWidget(&w.leftPanel)
	adder.AddWidget(&w.divider)
	adder.AddWidget(&w.rightPanel)

	w.backButton.SetText("← Back")
	w.backButton.OnDown(w.onBack)

	w.listContent.items = w.items
	w.listContent.selectedIdx = w.selectedItemIdx
	w.listContent.onSelected = func(itemIdx int) {
		w.selectedItemIdx = itemIdx
		w.listContent.selectedIdx = itemIdx
	}
	w.leftPanel.SetContent(&w.listContent)
	w.leftPanel.SetContentConstraints(basicwidget.PanelContentConstraintsFixedWidth)

	idx := w.selectedItemIdx
	if idx >= 0 && idx < len(w.items) {
		w.rightContent.entry = w.items[idx]
		w.rightPanel.SetContent(&w.rightContent)
	} else {
		w.rightPanel.SetContent(nil)
	}
	w.rightPanel.SetContentConstraints(basicwidget.PanelContentConstraintsFixedWidth)

	return nil
}

func (w *msgDetailWidget) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(ctx)

	w.headerItems = slices.Delete(w.headerItems, 0, len(w.headerItems))
	w.headerItems = append(w.headerItems,
		guigui.LinearLayoutItem{Widget: &w.backButton, Size: guigui.FixedSize(u * 3)},
	)
	header := guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     w.headerItems,
	}

	w.bodyItems = slices.Delete(w.bodyItems, 0, len(w.bodyItems))
	w.bodyItems = append(w.bodyItems,
		guigui.LinearLayoutItem{Widget: &w.leftPanel, Size: guigui.FlexibleSize(1)},
		guigui.LinearLayoutItem{Widget: &w.divider, Size: guigui.FixedSize(u / 2)},
		guigui.LinearLayoutItem{Widget: &w.rightPanel, Size: guigui.FlexibleSize(2)},
	)
	body := guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     w.bodyItems,
	}

	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	w.layoutItems = append(w.layoutItems,
		guigui.LinearLayoutItem{Layout: &header, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Layout: &body, Size: guigui.FlexibleSize(1)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Gap:       u / 2,
		Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
	}).LayoutWidgets(ctx, wb.Bounds(), layouter)
}
