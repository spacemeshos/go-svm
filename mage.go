//go:build ignore
// +build ignore

// See <https://magefile.org/zeroinstall/> for more information.

package main

import (
	"os"

	"github.com/magefile/mage/mage"
)

func main() { os.Exit(mage.Main()) }
