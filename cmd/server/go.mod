module main

go 1.22.5

replace memstorage => ./storage/memstorage
replace storage => ./storage

require (
	memstorage v0.0.0-00010101000000-000000000000 // indirect
	storage v0.0.0-00010101000000-000000000000 // indirect
)
