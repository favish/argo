# Developing locally

Clone this project and run main.go via `go run main.go`.  May be convienent to create an alias for this in your favorite shell's ~/.*rc file.

# Packaging binary

Create a new git tag for your release and run `make` from this directory.  Make is employed to automatically set the build and version information for the resulting binary.