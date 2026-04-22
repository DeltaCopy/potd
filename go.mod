module potd

go 1.25.8

replace potd/libs => ./libs

replace potd/models => ./models

require potd/libs v0.0.0-00010101000000-000000000000

require (
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	golang.org/x/sys v0.27.0 // indirect
	potd/models v0.0.0-00010101000000-000000000000 // indirect
)
