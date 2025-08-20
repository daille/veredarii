package util

/**
 *    Veredarii, software for interoperability.
 *    This file is part of Veredarii.
 *
 *    @author jcDaille
 *
 *
 *    MIT License
 *
 * Copyright (c) 2025 JC Daille
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

import (
	"Veredarii/general"
	"fmt"
)

var (
	Info    = Teal
	Debug   = White
	Warning = Yellow
	Error   = Red
	Fatal   = Red
)

var (
	White   = C("\033[1;37m%s\033[0m")
	Yellow  = C("\033[1;33m%s\033[0m")
	Teal    = C("\033[1;36m%s\033[0m")
	Red     = C("\033[1;31m%s\033[0m")
	Green   = C("\033[1;32m%s\033[0m")
	Purple  = C("\033[1;34m%s\033[0m")
	Magenta = C("\033[1;35m%s\033[0m")
	Black   = C("\033[1;30m%s\033[0m")
)

func C(base string) func(...interface{}) string {
	ret := func(args ...interface{}) string {
		if general.Configuration.Behavior.Log.Color {
			return fmt.Sprintf(base, fmt.Sprint(args...))
		} else {
			return fmt.Sprintf(fmt.Sprint(args...))
		}
	}
	return ret
}
