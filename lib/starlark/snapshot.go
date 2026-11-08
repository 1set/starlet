package starlark

import (
	"fmt"
	"math/big"
	"reflect"

	std "go.starlark.net/starlark"
)

const (
	maxDataDepth = 64
	maxDataNodes = 4096
	maxDataBytes = 1 << 20
	maxIntBits   = 4096
)

var integerBound = std.MakeBigInt(new(big.Int).Lsh(big.NewInt(1), maxIntBits))
var negativeIntegerBound = std.MakeBigInt(new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), maxIntBits)))

type objectKey struct {
	kind byte
	ptr  uintptr
	len  int
}

type dataCopy struct {
	thread *std.Thread
	nodes  int
	bytes  int
	active map[objectKey]bool
	memo   map[objectKey]std.Value
	height map[objectKey]int
}

func newDataCopy(thread *std.Thread) *dataCopy {
	return &dataCopy{thread: thread, active: make(map[objectKey]bool), memo: make(map[objectKey]std.Value), height: make(map[objectKey]int)}
}

func (c *dataCopy) chargeBytes(n int, path string) error {
	if n > maxDataBytes-c.bytes {
		return fmt.Errorf("%s: data exceeds %d bytes", path, maxDataBytes)
	}
	c.bytes += n
	return nil
}

func (c *dataCopy) value(value std.Value, path string, depth int) (std.Value, error) {
	c.nodes++ // Every edge counts, including an alias that hits the memo.
	if c.nodes > maxDataNodes {
		return nil, fmt.Errorf("%s: data exceeds %d nodes", path, maxDataNodes)
	}
	if depth > maxDataDepth {
		return nil, fmt.Errorf("%s: data exceeds depth %d", path, maxDataDepth)
	}
	if c.nodes%64 == 0 {
		if err := checkThread(c.thread); err != nil {
			return nil, err
		}
	}
	switch x := value.(type) {
	case std.NoneType, std.Bool, std.Float:
		return x, c.chargeBytes(8, path)
	case std.Int:
		upper, _ := x.Cmp(integerBound, 0)
		lower, _ := x.Cmp(negativeIntegerBound, 0)
		if upper >= 0 || lower <= 0 {
			return nil, fmt.Errorf("%s: integer exceeds %d bits", path, maxIntBits)
		}
		return x, c.chargeBytes((x.BigInt().BitLen()+7)/8, path)
	case std.String:
		return x, c.chargeBytes(len(x), path)
	case std.Bytes:
		return x, c.chargeBytes(len(x), path)
	case *std.List:
		if x == nil {
			return nil, fmt.Errorf("%s: nil list", path)
		}
		key := objectKey{kind: 'l', ptr: reflect.ValueOf(x).Pointer()}
		if known, found, err := c.lookup(key, path, depth); found || err != nil {
			return known, err
		}
		if x.Len() > maxDataNodes-c.nodes {
			return nil, fmt.Errorf("%s: list exceeds node limit", path)
		}
		out := std.NewList(nil)
		c.memo[key], c.active[key] = out, true
		defer delete(c.active, key)
		for i := 0; i < x.Len(); i++ {
			item, err := c.value(x.Index(i), fmt.Sprintf("%s[%d]", path, i), depth+1)
			if err != nil {
				return nil, err
			}
			_ = out.Append(item)
			c.childHeight(key, x.Index(i))
		}
		return out, nil
	case *std.Dict:
		if x == nil {
			return nil, fmt.Errorf("%s: nil dict", path)
		}
		key := objectKey{kind: 'd', ptr: reflect.ValueOf(x).Pointer()}
		if known, found, err := c.lookup(key, path, depth); found || err != nil {
			return known, err
		}
		if x.Len() > (maxDataNodes-c.nodes)/2 {
			return nil, fmt.Errorf("%s: dict exceeds node limit", path)
		}
		out := std.NewDict(x.Len())
		c.memo[key], c.active[key] = out, true
		defer delete(c.active, key)
		for i, pair := range x.Items() {
			k, err := c.value(pair[0], fmt.Sprintf("%s.key[%d]", path, i), depth+1)
			if err != nil {
				return nil, err
			}
			v, err := c.value(pair[1], fmt.Sprintf("%s.value[%d]", path, i), depth+1)
			if err != nil {
				return nil, err
			}
			if err := out.SetKey(k, v); err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			c.childHeight(key, pair[0])
			c.childHeight(key, pair[1])
		}
		return out, nil
	case std.Tuple:
		key := objectKey{kind: 't', ptr: reflect.ValueOf(x).Pointer(), len: len(x)}
		if known, found, err := c.lookup(key, path, depth); found || err != nil {
			return known, err
		}
		if len(x) > maxDataNodes-c.nodes {
			return nil, fmt.Errorf("%s: tuple exceeds node limit", path)
		}
		out := make(std.Tuple, len(x))
		c.memo[key], c.active[key] = out, true
		defer delete(c.active, key)
		for i, item := range x {
			v, err := c.value(item, fmt.Sprintf("%s[%d]", path, i), depth+1)
			if err != nil {
				return nil, err
			}
			out[i] = v
			c.childHeight(key, item)
		}
		return out, nil
	case *std.Set:
		if x == nil {
			return nil, fmt.Errorf("%s: nil set", path)
		}
		key := objectKey{kind: 's', ptr: reflect.ValueOf(x).Pointer()}
		if known, found, err := c.lookup(key, path, depth); found || err != nil {
			return known, err
		}
		if x.Len() > maxDataNodes-c.nodes {
			return nil, fmt.Errorf("%s: set exceeds node limit", path)
		}
		out := std.NewSet(x.Len())
		c.memo[key], c.active[key] = out, true
		defer delete(c.active, key)
		it := x.Iterate()
		defer it.Done()
		var item std.Value
		for i := 0; it.Next(&item); i++ {
			v, err := c.value(item, fmt.Sprintf("%s[%d]", path, i), depth+1)
			if err != nil {
				return nil, err
			}
			if err := out.Insert(v); err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			c.childHeight(key, item)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%s: unsupported data value %T", path, value)
	}
}

func (c *dataCopy) lookup(key objectKey, path string, depth int) (std.Value, bool, error) {
	if c.active[key] {
		return nil, false, fmt.Errorf("%s: reference cycle is unsupported", path)
	}
	v, found := c.memo[key]
	if found && depth+c.height[key] > maxDataDepth {
		return nil, false, fmt.Errorf("%s: aliased data exceeds depth %d", path, maxDataDepth)
	}
	return v, found, nil
}

func (c *dataCopy) childHeight(parent objectKey, value std.Value) {
	height := 1
	switch x := value.(type) {
	case *std.List:
		height += c.height[objectKey{kind: 'l', ptr: reflect.ValueOf(x).Pointer()}]
	case *std.Dict:
		height += c.height[objectKey{kind: 'd', ptr: reflect.ValueOf(x).Pointer()}]
	case *std.Set:
		height += c.height[objectKey{kind: 's', ptr: reflect.ValueOf(x).Pointer()}]
	case std.Tuple:
		height += c.height[objectKey{kind: 't', ptr: reflect.ValueOf(x).Pointer(), len: len(x)}]
	}
	if height > c.height[parent] {
		c.height[parent] = height
	}
}
