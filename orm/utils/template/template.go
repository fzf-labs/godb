package template

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

// DefaultTemplate is a tool to provides the text/template operations
type DefaultTemplate struct {
	text string
}

// NewTemplate returns an instance of defaultTemplate
func NewTemplate() *DefaultTemplate {
	return &DefaultTemplate{}
}

// Parse accepts a source template and returns defaultTemplate
func (t *DefaultTemplate) Parse(text string) *DefaultTemplate {
	t.text = text
	return t
}

// Execute returns the codes after the template executed
func (t *DefaultTemplate) Execute(data any) (*bytes.Buffer, error) {
	tem, err := template.New("").Parse(t.text)
	if err != nil {
		return nil, errors.Wrapf(err, "template parse error:%s", t.text)
	}

	buf := new(bytes.Buffer)
	if err = tem.Execute(buf, data); err != nil {
		return nil, errors.Wrapf(err, "template execute error:%s", t.text)
	}
	return buf, nil
}
