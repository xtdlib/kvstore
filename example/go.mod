module github.com/xtdlib/kvstore/example

go 1.24.3

require (
	github.com/xtdlib/kvstore v0.0.0
	github.com/xtdlib/rat v0.0.0-20250906071516-72ee30efa47e
	github.com/xtdlib/try v0.0.0-20250823064224-1656db820f45
)

require github.com/mattn/go-sqlite3 v1.14.32 // indirect

replace github.com/xtdlib/kvstore => ../
