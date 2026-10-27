// Package logger
package logger

type colors struct{}

var Color colors

func (colors) Reset() string  { return "\033[0m" }
func (colors) Red() string    { return "\033[31m" }
func (colors) Green() string  { return "\033[32m" }
func (colors) Yellow() string { return "\033[33m" }
func (colors) Blue() string   { return "\033[96m" }
