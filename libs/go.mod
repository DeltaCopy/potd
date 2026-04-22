module potd/libs

go 1.25.8

require (
	github.com/godbus/dbus/v5 v5.2.2
	potd/models v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.27.0 // indirect

replace potd/models => ../models
