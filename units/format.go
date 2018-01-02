package units

import (
	"fmt"
	"strconv"
)

type FmtOptions struct {
	Short     bool // if false, use unit shortname or symbol
	Precision int  // precision to truncate value
}

// ValueFormatter creates human-readable strings for a Unit value
type ValueFormatter func(Value, FmtOptions) string

func DefaultFormatter(v Value, opts FmtOptions) string {
	label := v.Unit.Name
	if opts.Short {
		label = v.Unit.Symbol
	}

	// make label plural if needed
	if v.Val > 1.0 {
		label = fmt.Sprintf("%ss", label)
	}

	vstr := strconv.FormatFloat(v.Val, 'f', opts.Precision, 64)

	return fmt.Sprintf("%s %s", vstr, label)
}