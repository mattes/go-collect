package file

import (
	"errors"
	"github.com/mattes/go-collect/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	var tests = []struct {
		testDesc string
		body     string
		label    string

		data   *data.Data
		labels []string
		err    error
	}{
		{
			testDesc: "no labels",
			body: `
      image: test
      name: foobar
      foo:
        - bar
        - hu
      `,
			label: "",

			data: data.ToData(map[string][]string{
				"image": []string{"test"},
				"name":  []string{"foobar"},
				"foo":   []string{"bar", "hu"},
			}),
			labels: []string{"default"},
			err:    nil,
		},

		{
			testDesc: "with label",
			body: `
      label:
        image: test
        name: foobar
      `,
			label: "label",
			data: data.ToData(map[string][]string{
				"image": []string{"test"},
				"name":  []string{"foobar"},
			}),
			labels: []string{"label"},
			err:    nil,
		},

		{
			testDesc: "get first label",
			body: `
      label:
        image: test
        name: foobar

      another-label:
        image: test

      label5:
        image: test1
      `,
			label: "",
			data: data.ToData(map[string][]string{
				"image": []string{"test"},
				"name":  []string{"foobar"},
			}),
			labels: []string{"label", "another-label", "label5"},
			err:    nil,
		},

		{
			testDesc: "get default label if exists",
			body: `
      label:
        image: test
        name: foobar

      default:
        image: use-this

      another-label: ~
      `,
			label: "",
			data: data.ToData(map[string][]string{
				"image": []string{"use-this"},
			}),
			labels: []string{"label", "default", "another-label"},
			err:    nil,
		},

		{
			testDesc: "only support two levels of indentation",
			body: `
      label:
        image: test
        name: foobar
        fails:
          foo: # ok
            - bar # not ok
      `,
			label:  "",
			data:   data.ToData(map[string][]string{}),
			labels: nil,
			err:    errors.New("only 2 levels of indentation allowed"),
		},
	}

	for _, tt := range tests {
		f := File{}
		f.body = []byte(tt.body)
		err := f.parse()
		assert.Equal(t, tt.err, err, tt.testDesc)
		if err == nil {
			assert.Equal(t, tt.data, f.getData(tt.label), tt.testDesc)
			assert.Equal(t, tt.labels, f.labels, tt.testDesc)
		}
	}
}

func TestSetPathFromUrl(t *testing.T) {
	// TODO
}

func TestSelectLabel(t *testing.T) {
	var tests = []struct {
		labelsInFile []string
		labelFromArg string
		selectLabel  string
		outLabel     string
	}{
		// all empty
		{[]string{}, "", "", ""},

		// use given label
		{[]string{"label"}, "label", "label", "label"},
		{[]string{"label"}, "", "label", "label"},

		// use default label
		{[]string{"default"}, "", "", "default"},

		// use first found label
		{[]string{"foo", "bar"}, "", "", "foo"},

		// always use default value if given
		{[]string{"foo", "default", "bar"}, "", "", "default"},
	}

	for _, tt := range tests {
		f := File{}
		f.label = tt.labelFromArg
		f.labels = tt.labelsInFile
		assert.Equal(t, tt.outLabel, f.selectLabel(tt.selectLabel))
	}
}
