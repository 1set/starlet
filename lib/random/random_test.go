package random_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/random"
	"go.starlark.net/starlark"
)

func TestLoadModule_Random(t *testing.T) {
	var (
		repeatTimes = 20
		one         = starlark.MakeInt(1)
		two         = starlark.MakeInt(2)
		three       = starlark.MakeInt(3)
	)
	tests := []struct {
		name        string
		script      string
		wantErr     string
		checkResult func(res starlark.Value) bool
	}{
		{
			name: `nil choice`,
			script: itn.HereDoc(`
				load('random', 'choice')
				choice()
			`),
			wantErr: `random.choice: missing argument for seq`,
		},
		{
			name: `no choice`,
			script: itn.HereDoc(`
				load('random', 'choice')
				choice([])
			`),
			wantErr: `cannot choose from an empty sequence`,
		},
		{
			name: `invalid choice`,
			script: itn.HereDoc(`
				load('random', 'choice')
				choice(123)
			`),
			wantErr: `random.choice: for parameter seq: got int, want starlark.Indexable`,
		},
		{
			name: `one choice`,
			script: itn.HereDoc(`
				load('random', 'choice')
				val = choice([1])
			`),
			checkResult: func(res starlark.Value) bool {
				return res.(starlark.Int) == one
			},
		},
		{
			name: `two choices`,
			script: itn.HereDoc(`
				load('random', 'choice')
				val = choice([1, 2])
			`),
			checkResult: func(res starlark.Value) bool {
				val := res.(starlark.Int)
				return val == one || val == two
			},
		},
		{
			name: `duplicate choices`,
			script: itn.HereDoc(`
				load('random', 'choice')
				val = choice([1, 1, 2, 3, 3, 2, 1])
			`),
			checkResult: func(res starlark.Value) bool {
				val := res.(starlark.Int)
				return val == one || val == two || val == three
			},
		},
		{
			name: `same choices`,
			script: itn.HereDoc(`
				load('random', 'choice')
				val = choice((3, 3, 3, 3, 3, 3, 3, 3, 3))
			`),
			checkResult: func(res starlark.Value) bool {
				return res.(starlark.Int) == three
			},
		},
		{
			name: "shuffle with invalid type",
			script: itn.HereDoc(`
				load('random', 'shuffle')
				x = 123
				shuffle(x)
			`),
			wantErr: `random.shuffle: for parameter seq: got int, want starlark.HasSetIndex`,
		},
		{
			name: "shuffle with immutable type",
			script: itn.HereDoc(`
				load('random', 'shuffle')
				x = (1, 2, 3)
				shuffle(x)
			`),
			wantErr: `random.shuffle: for parameter seq: got tuple, want starlark.HasSetIndex`,
		},
		{
			name: "shuffle with empty type",
			script: itn.HereDoc(`
				load('random', 'shuffle')
				val = []
				shuffle(val)
			`),
			checkResult: func(res starlark.Value) bool {
				return res.(*starlark.List).Len() == 0
			},
		},
		{
			name: "shuffle with one element",
			script: itn.HereDoc(`
				load('random', 'shuffle')
				val = [1]
				shuffle(val)
			`),
			checkResult: func(res starlark.Value) bool {
				l := res.(*starlark.List)
				return l.Len() == 1 && l.Index(0) == one
			},
		},
		{
			name: "shuffle with two elements",
			script: itn.HereDoc(`
				load('random', 'shuffle')
				val = [2, 3]
				shuffle(val)
			`),
			checkResult: func(res starlark.Value) bool {
				l := res.(*starlark.List)
				return l.Len() == 2 && ((l.Index(0) == two && l.Index(1) == three) || (l.Index(0) == three && l.Index(1) == two))
			},
		},
		{
			name: "shuffle with mutable type",
			script: itn.HereDoc(`
				load('random', 'shuffle')
				val = [1, 2, 3]
				shuffle(val)
				print(val)
			`),
			checkResult: func(res starlark.Value) bool {
				val := res.(*starlark.List)
				return val.Index(0) == one || val.Index(0) == two || val.Index(0) == three
			},
		},
		{
			name: "randint with less than 2 args",
			script: itn.HereDoc(`
				load('random', 'randint')
				randint()
			`),
			wantErr: `random.randint: missing argument for a`,
		},
		{
			name: "randint with more than 2 args",
			script: itn.HereDoc(`
				load('random', 'randint')
				randint(1, 2, 3)
			`),
			wantErr: `random.randint: got 3 arguments, want at most 2`,
		},
		{
			name: "randint with invalid type",
			script: itn.HereDoc(`
				load('random', 'randint')
				randint(1, '2')
			`),
			wantErr: `random.randint: for parameter b: got string, want int`,
		},
		{
			name: "randint with invalid range",
			script: itn.HereDoc(`
				load('random', 'randint')
				randint(2, 1)
			`),
			wantErr: `a must be less than or equal to b`,
		},
		{
			name: "randint with equal range",
			script: itn.HereDoc(`
				load('random', 'randint')
				val = randint(1, 1)
			`),
			checkResult: func(res starlark.Value) bool {
				return res.(starlark.Int) == one
			},
		},
		{
			name: "randint with range 1",
			script: itn.HereDoc(`
				load('random', 'randint')
				val = randint(1, 2)
			`),
			checkResult: func(res starlark.Value) bool {
				return res.(starlark.Int) == one || res.(starlark.Int) == two
			},
		},
		{
			name: "randint with range 2",
			script: itn.HereDoc(`
				load('random', 'randint')
				val = randint(1, 3)
				print(val)
			`),
			checkResult: func(res starlark.Value) bool {
				return res.(starlark.Int) == one || res.(starlark.Int) == two || res.(starlark.Int) == three
			},
		},
		{
			name: "randbytes with less than 1 args",
			script: itn.HereDoc(`
				load('random', 'randbytes')
				x = randbytes()
				assert.eq(len(x), 10)
			`),
		},
		{
			name: "randbytes with invalid args",
			script: itn.HereDoc(`
				load('random', 'randbytes')
				x = randbytes(-2)
				assert.eq(len(x), 10)
				y = randbytes(0)
				assert.eq(len(y), 10)
			`),
		},
		{
			name: "randbytes with invalid type",
			script: itn.HereDoc(`
				load('random', 'randbytes')
				randbytes('1')
			`),
			wantErr: `random.randbytes: for parameter n: got string, want int`,
		},
		{
			name: "randbytes with 1",
			script: itn.HereDoc(`
				load('random', 'randbytes')
				x = randbytes(1)
				assert.eq(len(x), 1)
				print(x)
			`),
		},
		{
			name: "randbytes with 20",
			script: itn.HereDoc(`
				load('random', 'randbytes')
				x = randbytes(20)
				assert.eq(len(x), 20)
			`),
		},
		{
			name: "random",
			script: itn.HereDoc(`
				load('random', 'random')
				val = random()
				print(val)
			`),
			checkResult: func(res starlark.Value) bool {
				f := res.(starlark.Float)
				return f >= 0 && f < 1
			},
		},
		{
			name: "uniform with less than 2 args",
			script: itn.HereDoc(`
				load('random', 'uniform')
				uniform()
			`),
			wantErr: `random.uniform: missing argument for a`,
		},
		{
			name: "uniform with more than 2 args",
			script: itn.HereDoc(`
				load('random', 'uniform')
				uniform(1, 2, 3)	
			`),
			wantErr: `random.uniform: got 3 arguments, want at most 2`,
		},
		{
			name: "uniform with invalid type",
			script: itn.HereDoc(`
				load('random', 'uniform')
				uniform('1', '2')
			`),
			wantErr: `random.uniform: for parameter a: got string, want float`,
		},
		{
			name: "uniform with int",
			script: itn.HereDoc(`
				load('random', 'uniform')
				val = uniform(1, 2)
			`),
			checkResult: func(res starlark.Value) bool {
				f := res.(starlark.Float)
				return f >= 1 && f < 2
			},
		},
		{
			name: "uniform with float",
			script: itn.HereDoc(`
				load('random', 'uniform')
				val = uniform(1.0, 2.0)
			`),
			checkResult: func(res starlark.Value) bool {
				f := res.(starlark.Float)
				return f >= 1 && f < 2
			},
		},
		{
			name: "uniform with equal range",
			script: itn.HereDoc(`
				load('random', 'uniform')
				val = uniform(1, 1)
			`),
			checkResult: func(res starlark.Value) bool {
				f := res.(starlark.Float)
				return f == 1
			},
		},
		{
			name: "uniform with reversed range",
			script: itn.HereDoc(`
				load('random', 'uniform')
				val = uniform(2, 1)
			`),
			checkResult: func(res starlark.Value) bool {
				f := res.(starlark.Float)
				return f >= 1 && f < 2
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 1; i <= repeatTimes; i++ {
				res, err := itn.ExecModuleWithErrorTest(t, random.ModuleName, random.LoadModule, tt.script, tt.wantErr)
				if (err != nil) != (tt.wantErr != "") {
					t.Errorf("random(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
					return
				}
				if tt.wantErr != "" {
					return
				}
				if tt.checkResult != nil && !tt.checkResult(res["val"]) {
					t.Errorf("random(%q) got unexpected result, actual result = %v", tt.name, res)
				}
			}
		})
	}
}
