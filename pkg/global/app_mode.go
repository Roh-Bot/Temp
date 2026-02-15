package global

import "flag"

var IsDevelopment bool

func ParseFlags() {
	value := flag.Bool("debug", false, "used to invoke debug mode")
	flag.Parse()
	IsDevelopment = *value
}
