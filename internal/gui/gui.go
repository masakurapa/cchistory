package gui

import (
	"image"

	"github.com/guigui-gui/guigui"
	_ "github.com/guigui-gui/guigui/basicwidget/cjkfont"
	"github.com/masakurapa/cchistory/internal/types"
)

type Widget interface {
	Build(context *guigui.Context, adder *guigui.ChildAdder) error
	Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter)
}

func Run(projectDir string, sessions []types.Session) error {
	return guigui.Run(newRoot(projectDir, sessions), &guigui.RunOptions{
		Title:         projectDir + " - cchistory",
		WindowMinSize: image.Pt(960, 600),
		AppScale:      1.5,
	})
}
