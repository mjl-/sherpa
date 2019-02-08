package main

import (
	"bitbucket.org/mjl/sherpa"
	"strings"
)

func sherpaDoc(sec *section) *sherpa.Doc {
	doc := &sherpa.Doc{
		Title:     sec.Name,
		Text:      sec.Text,
		Functions: []*sherpa.FunctionDoc{},
		Sections:  []*sherpa.Doc{},
		Types:     []sherpa.TypeDoc{},
	}
	for _, t := range sec.Types {
		tt := sherpa.TypeDoc{
			Name:   t.Name,
			Text:   t.Text,
			Fields: []sherpa.FieldDoc{},
		}
		for _, f := range t.Fields {
			ff := sherpa.FieldDoc{
				Name: f.Name,
				Text: f.Doc,
				Type: f.Type,
			}
			tt.Fields = append(tt.Fields, ff)
		}
		doc.Types = append(doc.Types, tt)
	}
	for _, fn := range sec.Functions {
		f := &sherpa.FunctionDoc{
			Name:   fn.Name,
			Text:   strings.TrimSpace(fn.Text),
			Params: fn.Params,
			Return: fn.Return,
		}
		doc.Functions = append(doc.Functions, f)
	}
	for _, subsec := range sec.Sections {
		doc.Sections = append(doc.Sections, sherpaDoc(subsec))
	}
	doc.Text = strings.TrimSpace(doc.Text)
	return doc
}
