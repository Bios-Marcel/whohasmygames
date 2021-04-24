package main

import (
	"fyne.io/fyne/v2"
)

// PrioVBoxLayout is a replacement for the fyne vbox, as it's barely usable
// due to the fact that you can't define whether a component should grow.
type PrioVBoxLayout struct {
	grow map[fyne.CanvasObject]bool
}

func NewPrioVBoxLayout() *PrioVBoxLayout {
	return &PrioVBoxLayout{
		grow: make(map[fyne.CanvasObject]bool),
	}
}

func (l *PrioVBoxLayout) SetGrow(object fyne.CanvasObject, grow bool) {
	l.grow[object] = grow
}

func (l *PrioVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 0)
}

func (l *PrioVBoxLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0)

	var objectsToGrow int
	//Space taken excluding the ones we want to grow anyways.
	var minHeightsVSpaceTaken float32
	for _, o := range objects {
		if l.grow[o] {
			objectsToGrow++
		} else {
			minHeightsVSpaceTaken += o.MinSize().Height
		}
	}

	spaceForEachGrowingComponent := (containerSize.Height - minHeightsVSpaceTaken) / float32(objectsToGrow)

	for _, o := range objects {
		if l.grow[o] {
			o.Resize(fyne.NewSize(containerSize.Width, spaceForEachGrowingComponent))
		} else {
			o.Resize(fyne.NewSize(containerSize.Width, o.MinSize().Height))
		}

		o.Move(pos)
		oSize := o.Size()
		pos = pos.Add(fyne.NewPos(0, oSize.Height))
	}
}
