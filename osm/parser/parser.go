package parser

import (
	"os"

	"github.com/missinglink/gosmparse"
	"github.com/sirupsen/logrus"
)

// Parser - PBF Parser
type Parser struct {
	file    *os.File
	decoder *gosmparse.Decoder
	logger  *logrus.Logger
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
	p.logger.Info("parsing started")
	err := p.decoder.Parse(handler, false)
	if err != nil {
		return err
	}
	p.logger.Info("parsing finished")
	return nil
}

// NewParser - Create a new parser for file at path
func NewParser(path string) (*Parser, error) {
	p := &Parser{logger: logrus.New()}
	err := p.open(path)
	if err != nil {
		return nil, err
	}

	return p, nil
}
