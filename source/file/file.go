package file

import (
	"errors"
	"fmt"
	"github.com/mattes/go-collect/data"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"path/filepath"
)

// File implements Source interface
type File struct {
	label string
	url   *url.URL
	path  string

	body []byte

	// TODO implement json, toml, ...
	yaml map[string]map[string][]string

	labels []string
}

func (s *File) Scheme() string {
	return "file"
}

func (s *File) ExampleUrl() string {
	return "file://config.yml"
}

func (s *File) Load(label string, u *url.URL) (*data.Data, error) {
	s.url = u
	s.label = label

	s.setPathFromUrl()

	if err := s.readFile(); err != nil {
		return nil, err
	}
	if err := s.parse(); err != nil {
		return nil, err
	}

	return s.getData(s.label), nil
}

func (s *File) setPathFromUrl() {
	// TODO what about windows and file://paths?

	if s.url.Host == "" {
		// assume absolute path
		// file:///home/config.yml
		s.path = s.url.Path
	} else {
		// assume relative path, this is not standard conform though
		// file://config.yml
		s.path = s.url.Host + "/" + s.url.Path
		s.path, _ = filepath.Abs(s.path)
	}
}

func (s *File) readFile() error {
	if s.path == "" {
		return errors.New("no file given")
	}
	body, err := ioutil.ReadFile(s.path)
	if err != nil {
		return err
	}
	s.body = body
	return nil
}

// parse parses the file content into a yaml struct
func (s *File) parse() error {

	// try to unmarshal with labels
	hasLabels := true
	var yamlWithLabels map[string]map[string]interface{}
	if err := yaml.Unmarshal(s.body, &yamlWithLabels); err != nil {
		// try without labels
		var yamlNoLabels map[string]interface{}
		if err := yaml.Unmarshal(s.body, &yamlNoLabels); err != nil {
			return errors.New("unable to parse yaml")
		}
		yamlWithLabels = map[string]map[string]interface{}{
			"default": yamlNoLabels,
		}
		hasLabels = false
	}

	// map[string]map[string]interface{} -> map[string]map[string][]string
	s.yaml = make(map[string]map[string][]string)
	for k, v := range yamlWithLabels {
		s.yaml[k] = make(map[string][]string)
		for k2, v2 := range v {
			s.yaml[k][k2] = make([]string, 0)
			switch v2.(type) {
			case []interface{}:
				s.yaml[k][k2] = interfaceSliceToStringSlice(v2.([]interface{}))

			case map[interface{}]interface{}:
				return errors.New("only 2 levels of indentation allowed")

			default:
				s.yaml[k][k2] = []string{fmt.Sprintf("%v", v2)}
			}
		}
	}

	// get labels and their order
	if hasLabels {
		var labelOrder yaml.MapSlice
		if err := yaml.Unmarshal(s.body, &labelOrder); err != nil {
			return errors.New("unable to parse yaml")
		}
		s.labels = make([]string, 0)
		for _, v := range labelOrder {
			s.labels = append(s.labels, v.Key.(string))
		}
	} else {
		s.labels = []string{"default"}
	}

	return nil
}

// labelExists returns bool if label exists in yaml
func (s *File) labelExists(label string) bool {
	for _, l := range s.labels {
		if l == label {
			return true
		}
	}
	return false
}

// selectLabel returns the label that should be used
func (s *File) selectLabel(label string) string {
	useLabel := ""

	if s.labelExists(label) {
		// use label
		useLabel = label

	} else if label == "" && s.labelExists(s.label) {
		// use default label
		useLabel = s.label

	} else if s.labelExists("default") {
		// use "default" label
		useLabel = "default"

	} else if len(s.labels) > 0 {
		// get first label found in file
		useLabel = s.labels[0]
	}

	return useLabel
}

// getYamlForLabel internally selects the right label to return
// given label (if exists), default label (if exist) or first label
func (s *File) getYamlForLabel(label string) (p map[string][]string, ok bool) {
	useLabel := s.selectLabel(label)
	for k, v := range s.yaml {
		if k == useLabel {
			return v, true
		}
	}
	return nil, false
}

// Data returns param.Data for given label
func (s *File) getData(label string) *data.Data {
	ps, ok := s.getYamlForLabel(s.selectLabel(label))
	if !ok {
		return data.New()
	}

	d := data.New()
	for k, v := range ps {
		d.Set(k, v...)
	}
	return d
}

// interfaceSliceToStringSlice converts:
// []interface{} -> []string
func interfaceSliceToStringSlice(in []interface{}) []string {
	out := make([]string, 0)
	for _, v := range in {
		out = append(out, fmt.Sprintf("%v", v))
	}
	return out
}
