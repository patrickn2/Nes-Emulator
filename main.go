package main

import "fmt"

func main() {
	fmt.Println("Nes Emulator in GoLang")
}

func a() {
	c()
}

func b() {
	a()
}

func c() {
	b()
}
