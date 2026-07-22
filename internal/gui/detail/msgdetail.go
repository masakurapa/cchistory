package detail

import (
	"image"
	"image/color"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	"github.com/masakurapa/cchistory/internal/types"
)

// coloredDivider draws a vertical line whose widget width provides left/right margin.
// Color and stroke width match basicwidget.Divider.
type coloredDivider struct {
	guigui.DefaultWidget
}

func (d *coloredDivider) Draw(ctx *guigui.Context, wb *guigui.WidgetBounds, dst *ebiten.Image) {
	b := wb.Bounds()
	strokeWidth := float32(1 * ctx.Scale())
	x := float32(b.Min.X+b.Max.X) / 2
	// same lightness as draw.DividerColor (light=0.8, dark=0.2)
	var clr color.RGBA
	if ctx.ColorMode() == ebiten.ColorModeDark {
		clr = color.RGBA{R: 51, G: 51, B: 51, A: 255}
	} else {
		clr = color.RGBA{R: 204, G: 204, B: 204, A: 255}
	}
	vector.StrokeLine(dst, x, float32(b.Min.Y), x, float32(b.Max.Y), strokeWidth, clr, false)
}

// turnListRow is a 2-line clickable row: date on top, first line of message below.
type turnListRow struct {
	guigui.DefaultWidget
	turn           types.Turn
	itemIdx        int
	selectedIdxPtr *int // points to turnListContent.selectedIdx for immediate Draw
	onSelected     func(itemIdx int)

	dateText basicwidget.Text
	msgText  basicwidget.Text
}

func (w *turnListRow) WriteStateKey(_ *guigui.Context, sw *guigui.StateKeyWriter) {
	selected := w.selectedIdxPtr != nil && *w.selectedIdxPtr == w.itemIdx
	sw.WriteBool(selected)
}

func (w *turnListRow) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	w.dateText.SetValue(w.turn.User.Timestamp.Local().Format("2006-01-02 15:04"))

	firstLine := w.turn.User.Content
	if i := strings.IndexByte(firstLine, '\n'); i >= 0 {
		firstLine = firstLine[:i]
	}
	w.msgText.SetValue(firstLine)

	adder.AddWidget(&w.dateText)
	adder.AddWidget(&w.msgText)
	return nil
}

func (w *turnListRow) Draw(_ *guigui.Context, wb *guigui.WidgetBounds, dst *ebiten.Image) {
	if w.selectedIdxPtr != nil && *w.selectedIdxPtr == w.itemIdx {
		dst.SubImage(wb.Bounds()).(*ebiten.Image).Fill(color.RGBA{R: 180, G: 180, B: 180, A: 80})
	}
}

func (w *turnListRow) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(ctx)
	b := wb.Bounds()
	hPad := u / 4
	mid := b.Min.Y + u
	layouter.LayoutWidget(&w.dateText, image.Rect(b.Min.X+hPad, b.Min.Y, b.Max.X-hPad, mid))
	layouter.LayoutWidget(&w.msgText, image.Rect(b.Min.X+hPad, mid, b.Max.X-hPad, b.Max.Y))
}

func (w *turnListRow) Measure(ctx *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(ctx)
	width, ok := constraints.FixedWidth()
	if !ok {
		width = 400
	}
	return image.Pt(width, u*2)
}

func (w *turnListRow) HandlePointingInput(_ *guigui.Context, wb *guigui.WidgetBounds) guigui.HandleInputResult {
	if wb.IsHitAtCursor() && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if w.onSelected != nil {
			w.onSelected(w.itemIdx)
		}
		return guigui.HandleInputByWidget(w)
	}
	return guigui.HandleInputResult{}
}

// compactBoundaryRow is a 2-line clickable row for compact boundary events.
type compactBoundaryRow struct {
	guigui.DefaultWidget
	cb             types.CompactBoundary
	itemIdx        int
	selectedIdxPtr *int
	onSelected     func(itemIdx int)

	dateText basicwidget.Text
	msgText  basicwidget.Text
}

func (w *compactBoundaryRow) WriteStateKey(_ *guigui.Context, sw *guigui.StateKeyWriter) {
	selected := w.selectedIdxPtr != nil && *w.selectedIdxPtr == w.itemIdx
	sw.WriteBool(selected)
}

func (w *compactBoundaryRow) Build(_ *guigui.Context, adder *guigui.ChildAdder) error {
	w.dateText.SetValue(w.cb.Timestamp.Local().Format("2006-01-02 15:04"))
	w.msgText.SetValue("auto compacting")
	adder.AddWidget(&w.dateText)
	adder.AddWidget(&w.msgText)
	return nil
}

func (w *compactBoundaryRow) Draw(_ *guigui.Context, wb *guigui.WidgetBounds, dst *ebiten.Image) {
	if w.selectedIdxPtr != nil && *w.selectedIdxPtr == w.itemIdx {
		dst.SubImage(wb.Bounds()).(*ebiten.Image).Fill(color.RGBA{R: 180, G: 180, B: 180, A: 80})
	}
}

func (w *compactBoundaryRow) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(ctx)
	b := wb.Bounds()
	hPad := u / 4
	mid := b.Min.Y + u
	layouter.LayoutWidget(&w.dateText, image.Rect(b.Min.X+hPad, b.Min.Y, b.Max.X-hPad, mid))
	layouter.LayoutWidget(&w.msgText, image.Rect(b.Min.X+hPad, mid, b.Max.X-hPad, b.Max.Y))
}

func (w *compactBoundaryRow) Measure(ctx *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(ctx)
	width, ok := constraints.FixedWidth()
	if !ok {
		width = 400
	}
	return image.Pt(width, u*2)
}

func (w *compactBoundaryRow) HandlePointingInput(_ *guigui.Context, wb *guigui.WidgetBounds) guigui.HandleInputResult {
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
	items       []types.TimelineItem
	selectedIdx int // index into items, -1 if none
	onSelected  func(int)

	rows         guigui.WidgetSlice[*turnListRow]
	compactRows  guigui.WidgetSlice[*compactBoundaryRow]
	dividers     guigui.WidgetSlice[*basicwidget.Divider]
	layoutItems  []guigui.LinearLayoutItem
}

func (w *turnListContent) rowCounts() (turns, compacts int) {
	for _, item := range w.items {
		if item.Turn != nil {
			turns++
		} else if item.CompactBoundary != nil {
			compacts++
		}
	}
	return
}

func (w *turnListContent) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	turns, compacts := w.rowCounts()
	total := turns + compacts
	w.rows.SetLen(turns)
	w.compactRows.SetLen(compacts)
	w.dividers.SetLen(max(0, total-1))

	ri, ci, di, total2 := 0, 0, 0, 0
	for i, item := range w.items {
		if item.Turn == nil && item.CompactBoundary == nil {
			continue
		}
		if total2 > 0 {
			adder.AddWidget(w.dividers.At(di))
			di++
		}
		if item.Turn != nil {
			row := w.rows.At(ri)
			row.turn = *item.Turn
			row.itemIdx = i
			row.selectedIdxPtr = &w.selectedIdx
			row.onSelected = w.onSelected
			adder.AddWidget(row)
			ri++
		} else {
			cr := w.compactRows.At(ci)
			cr.cb = *item.CompactBoundary
			cr.itemIdx = i
			cr.selectedIdxPtr = &w.selectedIdx
			cr.onSelected = w.onSelected
			adder.AddWidget(cr)
			ci++
		}
		total2++
	}
	return nil
}

func (w *turnListContent) layout(ctx *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(ctx)
	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	ri, ci, di, total := 0, 0, 0, 0
	for _, item := range w.items {
		if item.Turn == nil && item.CompactBoundary == nil {
			continue
		}
		if total > 0 {
			w.layoutItems = append(w.layoutItems,
				guigui.LinearLayoutItem{Widget: w.dividers.At(di), Size: guigui.FixedSize(1)},
			)
			di++
		}
		if item.Turn != nil {
			w.layoutItems = append(w.layoutItems,
				guigui.LinearLayoutItem{Widget: w.rows.At(ri)},
			)
			ri++
		} else {
			w.layoutItems = append(w.layoutItems,
				guigui.LinearLayoutItem{Widget: w.compactRows.At(ci)},
			)
			ci++
		}
		total++
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

// turnDetailWidget renders a single turn's full detail (user + assistant + meta).
type turnDetailWidget struct {
	guigui.DefaultWidget
	turn types.Turn

	userLabel   basicwidget.Text
	userArea    textAreaWidget
	assistLabel basicwidget.Text
	assistArea  textAreaWidget
	metaForm    metaFormWidget
	layoutItems []guigui.LinearLayoutItem
}

func (w *turnDetailWidget) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	w.userLabel.SetValue("User:")
	w.userArea.setText(w.turn.User.Content)
	adder.AddWidget(&w.userLabel)
	adder.AddWidget(&w.userArea)

	if !w.turn.Cancelled() {
		w.assistLabel.SetValue("Assistant:")
		w.assistArea.setText(w.turn.AssistantContent())
		adder.AddWidget(&w.assistLabel)
		adder.AddWidget(&w.assistArea)
	}

	w.metaForm.turn = w.turn
	adder.AddWidget(&w.metaForm)
	return nil
}

func (w *turnDetailWidget) layout(ctx *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(ctx)
	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	w.layoutItems = append(w.layoutItems,
		guigui.LinearLayoutItem{Widget: &w.userLabel, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Widget: &w.userArea},
	)
	if !w.turn.Cancelled() {
		w.layoutItems = append(w.layoutItems,
			guigui.LinearLayoutItem{Widget: &w.assistLabel, Size: guigui.FixedSize(u)},
			guigui.LinearLayoutItem{Widget: &w.assistArea},
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

func (w *turnDetailWidget) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	w.layout(ctx).LayoutWidgets(ctx, wb.Bounds(), layouter)
}

func (w *turnDetailWidget) Measure(ctx *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.layout(ctx).Measure(ctx, constraints)
}

// compactDetailWidget shows compact boundary details in the right panel.
type compactDetailWidget struct {
	guigui.DefaultWidget
	cb types.CompactBoundary

	summaryArea textAreaWidget
	metaForm    basicwidget.Form

	tsLabel      basicwidget.Text
	tsValue      basicwidget.Text
	triggerLabel basicwidget.Text
	triggerValue basicwidget.Text
	preLabel     basicwidget.Text
	preValue     basicwidget.Text
	postLabel    basicwidget.Text
	postValue    basicwidget.Text
	dropLabel    basicwidget.Text
	dropValue    basicwidget.Text

	formItems   []basicwidget.FormItem
	layoutItems []guigui.LinearLayoutItem
}

func (w *compactDetailWidget) Build(_ *guigui.Context, adder *guigui.ChildAdder) error {
	w.formItems = w.formItems[:0]
	row := func(label *basicwidget.Text, labelText string, value *basicwidget.Text, valueText string) {
		label.SetValue(labelText)
		value.SetValue(valueText)
		value.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
		value.SetSelectable(true)
		w.formItems = append(w.formItems, basicwidget.FormItem{PrimaryWidget: label, SecondaryWidget: value})
	}
	row(&w.tsLabel, "Timestamp", &w.tsValue, w.cb.Timestamp.Local().Format("2006-01-02 15:04:05"))
	if w.cb.Trigger != "" {
		row(&w.triggerLabel, "Trigger", &w.triggerValue, w.cb.Trigger)
	}
	row(&w.preLabel, "Pre Tokens", &w.preValue, formatTokens(w.cb.PreTokens))
	row(&w.postLabel, "Post Tokens", &w.postValue, formatTokens(w.cb.PostTokens))
	if w.cb.DroppedTokens > 0 {
		row(&w.dropLabel, "Dropped", &w.dropValue, formatTokens(w.cb.DroppedTokens))
	}
	w.metaForm.SetItems(w.formItems)
	adder.AddWidget(&w.metaForm)
	if w.cb.Summary != "" {
		w.summaryArea.setText(w.cb.Summary)
		adder.AddWidget(&w.summaryArea)
	}
	return nil
}

func (w *compactDetailWidget) layout(ctx *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(ctx)
	w.layoutItems = slices.Delete(w.layoutItems, 0, len(w.layoutItems))
	w.layoutItems = append(w.layoutItems, guigui.LinearLayoutItem{Widget: &w.metaForm})
	if w.cb.Summary != "" {
		w.layoutItems = append(w.layoutItems, guigui.LinearLayoutItem{Widget: &w.summaryArea})
	}
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     w.layoutItems,
		Gap:       u / 4,
		Padding:   guigui.Padding{Start: u / 2, Top: u / 2, End: u / 2, Bottom: u / 2},
	}
}

func (w *compactDetailWidget) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	w.layout(ctx).LayoutWidgets(ctx, wb.Bounds(), layouter)
}

func (w *compactDetailWidget) Measure(ctx *guigui.Context, constraints guigui.Constraints) image.Point {
	return w.layout(ctx).Measure(ctx, constraints)
}

type msgDetailWidget struct {
	guigui.DefaultWidget

	items           []types.TimelineItem
	selectedItemIdx int // index into items, -1 if none
	onBack          func(*guigui.Context)

	backButton      basicwidget.Button
	leftPanel       basicwidget.Panel
	listContent     turnListContent
	rightPanel     basicwidget.Panel
	rightContent   turnDetailWidget
	compactContent compactDetailWidget
	divider        coloredDivider

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
		item := w.items[idx]
		if item.Turn != nil {
			w.rightContent.turn = *item.Turn
			w.rightPanel.SetContent(&w.rightContent)
		} else if item.CompactBoundary != nil {
			w.compactContent.cb = *item.CompactBoundary
			w.rightPanel.SetContent(&w.compactContent)
		} else {
			w.rightPanel.SetContent(nil)
		}
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
