package ui

import "github.com/psychedelicdevx/bosun/internal/docker"

const standaloneKey = "(standalone)"

type row struct {
	header  bool
	project string
	ct      docker.Container
}

func (m Model) rows() []row {
	vis := m.visible()

	order := []string{}
	groups := map[string][]docker.Container{}
	for _, c := range vis {
		p := c.Project
		if p == "" {
			p = standaloneKey
		}
		if _, ok := groups[p]; !ok {
			order = append(order, p)
		}
		groups[p] = append(groups[p], c)
	}

	if len(order) == 1 && order[0] == standaloneKey {
		rs := make([]row, 0, len(vis))
		for _, c := range groups[standaloneKey] {
			rs = append(rs, row{ct: c})
		}
		return rs
	}

	var rs []row
	for _, p := range order {
		rs = append(rs, row{header: true, project: p})
		if m.collapsed[p] {
			continue
		}
		for _, c := range groups[p] {
			rs = append(rs, row{ct: c})
		}
	}
	return rs
}

func (m Model) currentRow() (row, bool) {
	rs := m.rows()
	if m.cursor < 0 || m.cursor >= len(rs) {
		return row{}, false
	}
	return rs[m.cursor], true
}
