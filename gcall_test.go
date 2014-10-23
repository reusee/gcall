package gcall

import (
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	Require("Gtk", "3.0")
	pt("major version %v\n", Call("Gtk.get_major_version"))

	Call("Gtk.init", 0, nil)
	win := New("Gtk.Window", 0)

	grid := New("Gtk.Grid")
	Call("Gtk.Container.add", win, grid)

	button := New("Gtk.Button")
	Call("Gtk.Button.set_label", button, "买买买")
	Call("Gtk.Container.add", grid, button)
	Connect(button, "clicked", func() {
		pt("clicked %v\n", time.Now())
	})

	Call("Gtk.Widget.show_all", win)
	Connect(win, "destroy", func() {
		Call("Gtk.main_quit")
	})
	Call("Gtk.main")
}
