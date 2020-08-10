package prometheus

import "strings"

type PromQLHandler struct {
	Prom string
}

func NewPromQL(prom string) *PromQLHandler {
	return &PromQLHandler{Prom: prom}
}

func (p *PromQLHandler) Replace(name, value string) *PromQLHandler {
	p.Prom = strings.Replace(p.Prom, "$"+name, value, -1)
	return p
}

func (p *PromQLHandler) ReplaceAll(labelMaps map[string]string) *PromQLHandler {
	for k, v := range labelMaps {
		p.Prom = strings.Replace(p.Prom, "$"+k, v, -1)
	}
	return p
}

func (p *PromQLHandler) GetValue() string {
	return p.Prom
}
