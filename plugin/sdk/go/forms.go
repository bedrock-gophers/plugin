package guest

// MenuFormData describes a simple menu form that can be shown to a player.
// The form is fire-and-forget: submit/close responses are ignored.
type MenuFormData struct {
	Title   string
	Body    string
	Buttons []string
}

// ModalFormData describes a simple two-button modal form.
// The form is fire-and-forget: submit/close responses are ignored.
type ModalFormData struct {
	Title   string
	Body    string
	Confirm string
	Cancel  string
}
