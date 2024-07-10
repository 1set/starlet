package stats_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/stats"
)

func TestLoadModule_Stats(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: "Euclidean Distance",
			script: itn.HereDoc(`
				load('stats', 'euclidean_distance')
				assert.eq(euclidean_distance([3, 4], [0, 0]), 5.0)
				assert.eq(euclidean_distance([1, 1], [4, 5]), 5.0)
			`),
		},
		{
			name: "Manhattan Distance",
			script: itn.HereDoc(`
				load('stats', 'manhattan_distance')
				assert.eq(manhattan_distance([3, 4], [0, 0]), 7.0)
				assert.eq(manhattan_distance([1, 1], [4, 5]), 7.0)
			`),
		},
		{
			name: "Softmax",
			script: itn.HereDoc(`
				load('stats', 'softmax')
				assert.eq(softmax([1, 1, 1]), [0.3333333333333333, 0.3333333333333333, 0.3333333333333333])
				assert.eq(softmax([1, 2, 3]), [0.09003057317038046, 0.24472847105479764, 0.6652409557748218])
			`),
		},
		{
			name: "Invalid Softmax",
			script: itn.HereDoc(`
				load('stats', 'softmax')
				softmax(True)
			`),
			wantErr: `softmax: for parameter 1: got bool, want iterable`,
		},
		{
			name: "Wrong Softmax",
			script: itn.HereDoc(`
				load('stats', 'softmax')
				softmax([])
			`),
			wantErr: `Input must not be empty.`,
		},
		{
			name: "Sigmoid",
			script: itn.HereDoc(`
				load('stats', 'sigmoid')
				assert.eq(sigmoid([0, 2, 4]), [0.5, 0.8807970779778823, 0.9820137900379085])
				assert.eq(sigmoid([-2, -2, -2]), [0.11920292202211755, 0.11920292202211755, 0.11920292202211755])
			`),
		},
		{
			name: "Mode",
			script: itn.HereDoc(`
				load('stats', 'mode')
				assert.eq(mode([1, 2, 2, 3, 4]), [2.0])
				assert.eq(mode([1, 1, 2, 3, 3]), [1.0, 3.0])
			`),
		},
		{
			name: "Empty Mode",
			script: itn.HereDoc(`
				load('stats', 'mode')
				mode([])
			`),
			wantErr: `Input must not be empty.`,
		},
		{
			name: "Sum",
			script: itn.HereDoc(`
				load('stats', 'sum')
				assert.eq(sum([1, 2, 3, 4]), 10.0)
			`),
		},
		{
			name: "Invalid Sum",
			script: itn.HereDoc(`
				load('stats', 'sum')
				sum([])
			`),
			wantErr: `Input must not be empty.`,
		},
		{
			name: "Max",
			script: itn.HereDoc(`
				load('stats', 'max')
				assert.eq(max([1, 2, 3, 4]), 4.0)
				assert.eq(max([-1, -2, -3, -4]), -1.0)
			`),
		},
		{
			name: "Min",
			script: itn.HereDoc(`
				load('stats', 'min')
				assert.eq(min([1, 2, 3, 4]), 1.0)
				assert.eq(min([-1, -2, -3, -4]), -4.0)
			`),
		},
		{
			name: "Mean",
			script: itn.HereDoc(`
				load('stats', 'mean')
				assert.eq(mean([1, 2, 3, 4]), 2.5)
				assert.eq(mean([0, 0, 0, 0]), 0.0)
			`),
		},
		{
			name: "Geometric Mean",
			script: itn.HereDoc(`
				load('stats', 'geometric_mean')
				assert.eq(geometric_mean([1, 2, 3, 4]), 2.213363839400643)
				assert.eq(geometric_mean([1, 1, 1, 1]), 1.0)
			`),
		},
		{
			name: "Harmonic Mean",
			script: itn.HereDoc(`
				load('stats', 'harmonic_mean')
				assert.eq(harmonic_mean([1, 1, 1, 1]), 1.0)
				assert.eq(harmonic_mean([1, 2, 3, 6]), 2.0)
			`),
		},
		{
			name: "Trimean",
			script: itn.HereDoc(`
				load('stats', 'trimean')
				assert.eq(trimean([1, 2, 3, 4, 5]), 3.0)
				assert.eq(trimean([1, 3, 5, 6, 7, 8, 12]), 5.75)
			`),
		},
		{
			name: "Median",
			script: itn.HereDoc(`
				load('stats', 'median')
				assert.eq(median([1, 2, 3, 4, 5]), 3.0)
				assert.eq(median([1, 3, 3, 6, 7]), 3.0)
			`),
		},
		{
			name: "Percentile",
			script: itn.HereDoc(`
				load('stats', 'percentile')
				assert.eq(percentile([1, 2, 3, 4, 5], 40), 2)
				assert.eq(percentile([1, 2, 3, 4, 5], 50), 2.5)
				assert.eq(percentile([1, 2, 3, 4, 5], 90), 4.5)
			`),
		},
		{
			name: "Invalid Percentile",
			script: itn.HereDoc(`
				load('stats', 'percentile')
				percentile([1, 2, 3, 4, 5])
			`),
			wantErr: `percentile: got 1 arguments, want 2`,
		},
		{
			name: "Wrong Percentile",
			script: itn.HereDoc(`
				load('stats', 'percentile')
				percentile([1, 2, 3, 4, 5], -10)
			`),
			wantErr: `Input is outside of range.`,
		},
		{
			name: "Variance",
			script: itn.HereDoc(`
				load('stats', 'variance')
				assert.eq(variance([1, 2, 3, 4, 5]), 2.0)
				assert.eq(variance([1, 1, 1, 1]), 0.0)
			`),
		},
		{
			name: "Covariance",
			script: itn.HereDoc(`
				load('stats', 'covariance')
				assert.eq(covariance([1, 2, 3], [4, 5, 6]), 1.0)
				assert.eq(covariance([1, 1, 1], [1, 1, 1]), 0.0)
			`),
		},
		{
			name: "Wrong Covariance",
			script: itn.HereDoc(`
				load('stats', 'covariance')
				covariance([1, 2, 3], [4, 5])
			`),
			wantErr: `Must be the same length.`,
		},
		{
			name: "Sample Variance",
			script: itn.HereDoc(`
				load('stats', 'sample_variance')
				assert.eq(sample_variance([1, 2, 3, 4, 5]), 2.5)
				assert.eq(sample_variance([1, 1, 1, 1]), 0.0)
			`),
		},
		{
			name: "Standard Deviation",
			script: itn.HereDoc(`
				load('stats', 'standard_deviation')
				assert.eq(standard_deviation([1, 2, 3, 4, 5]), 1.4142135623730951)
				assert.eq(standard_deviation([1, 1, 1, 1]), 0.0)
			`),
		},
		{
			name: "Sample",
			script: itn.HereDoc(`
				load('stats', 'sample')
				r1 = sample(data=[1, 2, 3, 4], take=3, replace=False)
				assert.eq(len(r1), 3)
				r2 = sample(data=[1, 2, 3, 4], take=5, replace=True)
				assert.eq(len(r2), 5)
				print(r1)
				print(r2)
			`),
		},
		{
			name: "Invalid Sample",
			script: itn.HereDoc(`
				load('stats', 'sample')
				sample([1,2,3])
			`),
			wantErr: `sample: missing argument for take`,
		},
		{
			name: "Wrong Sample",
			script: itn.HereDoc(`
				load('stats', 'sample')
				sample([1,2,3,4], 5, False)
			`),
			wantErr: `Input is outside of range.`,
		},
		{
			name: "Correlation",
			script: itn.HereDoc(`
				load('stats', 'correlation')
				assert.eq(correlation([1, 2, 3], [1, 2, 3]), 1.0)
				assert.eq(correlation([1, 2, 3], [6, 5, 4]), -1.0)
			`),
		},
		{
			name: "Invalid Correlation",
			script: itn.HereDoc(`
				load('stats', 'correlation')
				correlation([1,2,3])
			`),
			wantErr: `correlation: got 1 arguments, want 2`,
		},
		{
			name: "Pearson Correlation",
			script: itn.HereDoc(`
				load('stats', 'pearson')
				assert.eq(pearson([1, 2, 3], [1, 2, 3]), 1.0)
				assert.eq(pearson([1, 2, 3], [6, 5, 4]), -1.0)
			`),
		},
		{
			name: "Invalid Argument Count",
			script: itn.HereDoc(`
				load('stats', 'mean')
				mean([1, 2, 3], [4])
			`),
			wantErr: "mean: got 2 arguments, want 1",
		},
		{
			name: "Invalid Input Type",
			script: itn.HereDoc(`
				load('stats', 'mean')
				mean("1, 2, 3")
				assert.fail("should not reach here")
			`),
			wantErr: "mean: for parameter 1: got string, want iterable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, stats.ModuleName, stats.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("stats(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
