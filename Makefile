include $(GOROOT)/src/Make.inc

TARG=gobook
GOFILES=\
	src/main.go\
	src/mustache.go\
	src/uuid4.go\
	src/db.go\
	src/utils.go

include $(GOROOT)/src/Make.cmd