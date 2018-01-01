package units

import (
	"fmt"

	valuate "github.com/Knetic/govaluate"
	"github.com/bcicen/xiny/bfstree"
)

type ConversionFn func(float64) float64

type Conversion struct {
	from Unit
	to   Unit
	Fn   ConversionFn
}

// Conversion implements bfstree.Edge interface
func (c Conversion) To() string   { return c.to.Name }
func (c Conversion) From() string { return c.from.Name }

type Quantity struct {
	Name      string
	Units     []Unit
	Formatter ValueFormatter
	conv      []Conversion
}

func NewQuantity(name string, formatter ValueFormatter) *Quantity {
	if _, ok := All[name]; !ok {
		All[name] = &Quantity{
			Name:      name,
			Formatter: formatter,
		}
	}
	fmt.Printf("added new quantity %s\n", name)
	return All[name]
}

// Create a new Unit within this quantity and return it
func (q *Quantity) NewUnit(name, symbol string) Unit {
	u := Unit{name, symbol}
	q.Units = append(q.Units, u)
	fmt.Printf("added new unit %s\n", name)
	return u
}

// return unit matching name or symbol provided
func (q *Quantity) FindUnit(s string) (Unit, error) {
	for _, u := range q.Units {
		if u.Name == s || u.Symbol == s {
			return u, nil
		}
	}
	return Unit{}, fmt.Errorf("unit \"%s\"not found", s)
}

// Create a conversion and the inverse, given a ratio of from Unit
// in to Unit
func (q *Quantity) NewRatioConv(from, to Unit, ratio float64) {
	q.NewConv(from, to, fmt.Sprintf("x * %.12f", ratio))
	q.NewConv(to, from, fmt.Sprintf("x / %.12f", ratio))
}

func (q *Quantity) NewConv(from, to Unit, formula string) {
	expr, err := valuate.NewEvaluableExpression(formula)
	if err != nil {
		panic(err)
	}

	// create conversion function
	fn := func(x float64) float64 {
		params := make(map[string]interface{})
		params["x"] = x

		res, err := expr.Evaluate(params)
		if err != nil {
			panic(err)
		}
		return res.(float64)
	}

	q.conv = append(q.conv, Conversion{from, to, fn})
}

func (q *Quantity) FmtValue(v Value, opts FmtOptions) string { return q.Formatter(v, opts) }

// Convert provided value from one unit to another
func (q *Quantity) Convert(v Value, to Unit) (newVal Value, err error) {
	fns, err := q.resolve(v.Unit, to)
	if err != nil {
		return newVal, err
	}

	fVal := v.Val
	for _, fn := range fns {
		fVal = fn(fVal)
	}

	return Value{fVal, to}, nil
}

// resolve a path of one or more conversions between two units
func (q *Quantity) resolve(from, to Unit) (fns []ConversionFn, err error) {
	tree := bfstree.NewBFSTree()
	for _, cnv := range q.conv {
		tree.AddEdge(cnv)
	}

	path, err := tree.FindPath(from.Name, to.Name)
	if err != nil {
		return fns, err
	}

	for _, edge := range path.Edges() {
		fmt.Printf("%s -> %s\n", edge.From(), edge.To())
		fn, err := q.lookup(edge.From(), edge.To())
		if err != nil {
			return fns, err
		}
		fns = append(fns, fn)
	}

	return fns, nil
}

// find conversion function between two units
func (q *Quantity) lookup(from, to string) (ConversionFn, error) {
	for _, c := range q.conv {
		if c.From() == from && c.To() == to {
			return c.Fn, nil
		}
	}
	return nil, fmt.Errorf("conversion not found")
}
