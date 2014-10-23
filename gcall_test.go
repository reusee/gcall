package gcall

import (
	"testing"
	"unsafe"
)

func TestAll(t *testing.T) {
	Require("Gtk", "3.0")
	pt("major version %v\n", Call("Gtk.get_major_version"))

	Call("Gtk.init", 0, nil)
	win := Call("Gtk.Window.new", 0)[0].(unsafe.Pointer)
	button := Call("Gtk.Button.new")[0].(unsafe.Pointer)
	Call("Gtk.Button.set_label", button, "买买买")
	Call("Gtk.Container.add", win, button)
	Call("Gtk.Widget.show_all", win)
	Connect(win, "destroy", func() {
		Call("Gtk.main_quit")
	})
	Call("Gtk.main")
}
