package clingy

import (
	"fmt"
	"reflect"
)

type paramOpts struct {
	opt   bool
	rep   bool
	short byte
	adv   bool
	typ   string
	fns   []interface{}
}

type param struct {
	paramOpts
	name string
	def  interface{}
	desc string
	typ  reflect.Type
	err  error
}

func (p *param) zeroType() reflect.Type {
	typ := p.typ
	if p.opt && !p.rep {
		typ = reflect.PtrTo(typ)
	} else if p.rep {
		typ = reflect.SliceOf(typ)
	}
	return typ
}

func (p *param) zero() interface{} {
	return zero(p.zeroType())
}

func (p *param) flagType() string {
	if p.paramOpts.typ != "" {
		return p.paramOpts.typ
	}
	switch p.typ {
	case boolType:
		return ""
	case durationType:
		return "duration"
	default:
		return p.typ.Name()
	}
}

type charSet [256 / 32]uint32

func (c *charSet) Set(x byte)      { c[x/32] |= 1 << (x % 32) }
func (c *charSet) Has(x byte) bool { return c[x/32]&(1<<(x%32)) != 0 }

type params struct {
	list   []*param
	set    map[string]*param
	shorts charSet
}

func newParams() *params {
	return &params{
		set: make(map[string]*param),
	}
}

func (ps *params) count() int { return len(ps.list) }

func (ps *params) hasErrors() bool {
	for _, p := range ps.list {
		if p.err != nil {
			return true
		}
	}
	return false
}

func (ps *params) iter(cb func(*param)) {
	for _, p := range ps.list {
		cb(p)
	}
}

func (ps *params) newParam(name, desc string, def interface{}, options ...Option) *param {
	p := &param{name: name, def: def, desc: desc}
	for _, opt := range options {
		opt.do(&p.paramOpts)
	}
	if _, ok := ps.set[name]; ok {
		panic(fmt.Sprintf("parameter already defined with name: %q", name))
	} else if p.short != 0 && ps.shorts.Has(p.short) {
		panic(fmt.Sprintf("parameter already defined with short-name: %q", p.short))
	}
	var err error
	p.typ, err = checkFns(p.fns)
	if err != nil {
		panic(fmt.Sprintf("parameter has invalid transformation functions: %v", err))
	}
	ps.list = append(ps.list, p)
	ps.set[name] = p
	ps.shorts.Set(p.short)
	return p
}
