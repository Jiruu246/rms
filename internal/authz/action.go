package authz

// Action names an operation being authorized, in "resource:verb" form, e.g.
// "menu_item:update". Declare Action constants next to the resource they
// govern (in that resource's service file), not here — this file stays
// generic. None are declared yet; the first resource wired to this
// framework should add its own.
type Action string
