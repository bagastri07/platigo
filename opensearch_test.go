package platigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpensearchClient(t *testing.T) {
	type args struct {
		config *OSConfig
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal",
			args: args{
				config: &OSConfig{
					Addresses:          []string{"localhost:9200"},
					InsecureSkipVerify: true,
					Username:           "",
					Password:           "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOpensearchClient(tt.args.config)
			assert.NotNil(t, got)
			assert.NoError(t, err)
		})
	}
}
