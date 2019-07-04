package parser

import (
	"os"

	"github.com/missinglink/gosmparse"
)

// Parser - PBF Parser
type Parser struct {
	file    *os.File
	decoder *gosmparse.Decoder
}

// open - open file path
func (p *Parser) open(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	p.file = file
	p.decoder = gosmparse.NewDecoder(file)
	return nil
}

// Parse - execute parser
func (p *Parser) Parse(handler gosmparse.OSMReader) error {
	err := p.decoder.Parse(handler, false)
	if err != nil {
		return err
	}
	return nil
}

// NewParser - Create a new parser for file at path
func NewParser(path string) (*Parser, error) {
	p := &Parser{}
	err := p.open(path)
	if err != nil {
		return nil, err
	}

	return p, nil
}
