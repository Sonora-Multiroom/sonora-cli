// Package exitcode is a compile-only stub; T018 replaces it with the real
// mapping from hub.ErrorClass to process exit codes.
package exitcode

import "github.com/Sonora-Multiroom/sonora-cli/hub"

const Usage = 0

func For(c hub.ErrorClass) int { return -1 }
