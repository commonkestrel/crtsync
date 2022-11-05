package padding

import "fmt"

const (
    RIGHT = iota
    LEFT
    EDGES
)

func Pad(in any, size int, direction uint8) string {
    base := fmt.Sprint(in)
    if size < len(base) {
        return base
    }
    padding := size-len(base)

    switch direction {
    case RIGHT:
        var space string
        for i := 0; i < padding; i++ {
            space += " "
        }
        return base + space
    case LEFT:
        var space string
        for i := 0; i < padding; i++ {
            space += " "
        }
        return space + base
    case EDGES:
        half := padding/2
        var space string
        for i := 0; i < half; i++ {
            space += " "
        }
        base = space + base

        space = ""
        for i := 0; i < padding-half; i++ {
            space += " "
        }
        return base + space
    default:
        return base
    }
}

func Fill(in any, size int, fillchar rune, direction uint8) string {
    base := fmt.Sprint(in)
    if size < len(base) {
        return base
    }
    padding := size-len(base)
    fill := string(fillchar)


    switch direction {
    case RIGHT:
        var space string
        for i := 0; i < padding; i++ {
            space += fill
        }
        return base + space
    case LEFT:
        var space string
        for i := 0; i < padding; i++ {
            space += fill
        }
        return space + base
    case EDGES:
        half := padding/2
        var space string
        for i := 0; i < half; i++ {
            space += fill
        }
        base = space + base

        space = ""
        for i := 0; i < padding-half; i++ {
            space += fill
        }
        return base + space
    default:
        return base
    }
}