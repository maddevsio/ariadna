// Package osmpbf decodes OpenStreetMap (OSM) PBF files.
// Use this package by creating a NewDecoder and passing it a PBF file.
// Use Start to start decoding process.
// Use Decode to return Node, Way and Relation structs.
package osmpbf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/qedus/osmpbf/OSMPBF"
)

const (
	maxBlobHeaderSize = 64 * 1024

	initialBlobBufSize = 1 * 1024 * 1024

	// MaxBlobSize is maximum supported blob size.
	MaxBlobSize = 32 * 1024 * 1024
)

var (
	parseCapabilities = map[string]bool{
		"OsmSchema-V0.6": true,
		"DenseNodes":     true,
	}
)

type BoundingBox struct {
	Left   float64
	Right  float64
	Top    float64
	Bottom float64
}

type Header struct {
	BoundingBox                      *BoundingBox
	RequiredFeatures                 []string
	OptionalFeatures                 []string
	WritingProgram                   string
	Source                           string
	OsmosisReplicationTimestamp      time.Time
	OsmosisReplicationSequenceNumber int64
	OsmosisReplicationBaseUrl        string
}

type Info struct {
	Version   int32
	Uid       int32
	Timestamp time.Time
	Changeset int64
	User      string
	Visible   bool
}

type Node struct {
	ID   int64
	Lat  float64
	Lon  float64
	Tags map[string]string
	Info Info
}

type Way struct {
	ID      int64
	Tags    map[string]string
	NodeIDs []int64
	Info    Info
}

type Relation struct {
	ID      int64
	Tags    map[string]string
	Members []Member
	Info    Info
}

type MemberType int

const (
	NodeType MemberType = iota
	WayType
	RelationType
)

type Member struct {
	ID   int64
	Type MemberType
	Role string
}

type pair struct {
	i interface{}
	e error
}

// A Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type Decoder struct {
	r          io.Reader
	serializer chan pair

	buf *bytes.Buffer

	// store header block
	header *Header
	// synchronize header deserialization
	headerOnce sync.Once

	// for data decoders
	inputs  []chan<- pair
	outputs []<-chan pair
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	d := &Decoder{
		r:          r,
		serializer: make(chan pair, 8000), // typical PrimitiveBlock contains 8k OSM entities
	}
	d.SetBufferSize(initialBlobBufSize)
	return d
}

// SetBufferSize sets initial size of decoding buffer. Default value is 1MB, you can set higher value
// (for example, MaxBlobSize) for (probably) faster decoding, or lower value for reduced memory consumption.
// Any value will produce valid results; buffer will grow automatically if required.
func (dec *Decoder) SetBufferSize(n int) {
	dec.buf = bytes.NewBuffer(make([]byte, 0, n))
}

// Header returns file header.
func (dec *Decoder) Header() (*Header, error) {
	// deserialize the file header
	return dec.header, dec.readOSMHeader()
}

// Start decoding process using n goroutines.
func (dec *Decoder) Start(n int) error {
	if n < 1 {
		n = 1
	}

	if err := dec.readOSMHeader(); err != nil {
		return err
	}

	// start data decoders
	for i := 0; i < n; i++ {
		input := make(chan pair)
		output := make(chan pair)
		go func() {
			dd := new(dataDecoder)
			for p := range input {
				if p.e == nil {
					// send decoded objects or decoding error
					objects, err := dd.Decode(p.i.(*OSMPBF.Blob))
					output <- pair{objects, err}
				} else {
					// send input error as is
					output <- pair{nil, p.e}
				}
			}
			close(output)
		}()

		dec.inputs = append(dec.inputs, input)
		dec.outputs = append(dec.outputs, output)
	}

	// start reading OSMData
	go func() {
		var inputIndex int
		for {
			input := dec.inputs[inputIndex]
			inputIndex = (inputIndex + 1) % n

			blobHeader, blob, err := dec.readFileBlock()
			if err == nil && blobHeader.GetType() != "OSMData" {
				err = fmt.Errorf("unexpected fileblock of type %s", blobHeader.GetType())
			}
			if err == nil {
				// send blob for decoding
				input <- pair{blob, nil}
			} else {
				// send input error as is
				input <- pair{nil, err}
				for _, input := range dec.inputs {
					close(input)
				}
				return
			}
		}
	}()

	go func() {
		var outputIndex int
		for {
			output := dec.outputs[outputIndex]
			outputIndex = (outputIndex + 1) % n

			p := <-output
			if p.i != nil {
				// send decoded objects one by one
				for _, o := range p.i.([]interface{}) {
					dec.serializer <- pair{o, nil}
				}
			}
			if p.e != nil {
				// send input or decoding error
				dec.serializer <- pair{nil, p.e}
				close(dec.serializer)
				return
			}
		}
	}()

	return nil
}

// Decode reads the next object from the input stream and returns either a
// pointer to Node, Way or Relation struct representing the underlying OpenStreetMap PBF
// data, or error encountered. The end of the input stream is reported by an io.EOF error.
//
// Decode is safe for parallel execution. Only first error encountered will be returned,
// subsequent invocations will return io.EOF.
func (dec *Decoder) Decode() (interface{}, error) {
	p, ok := <-dec.serializer
	if !ok {
		return nil, io.EOF
	}
	return p.i, p.e
}

func (dec *Decoder) readFileBlock() (*OSMPBF.BlobHeader, *OSMPBF.Blob, error) {
	blobHeaderSize, err := dec.readBlobHeaderSize()
	if err != nil {
		return nil, nil, err
	}

	blobHeader, err := dec.readBlobHeader(blobHeaderSize)
	if err != nil {
		return nil, nil, err
	}

	blob, err := dec.readBlob(blobHeader)
	if err != nil {
		return nil, nil, err
	}

	return blobHeader, blob, err
}

func (dec *Decoder) readBlobHeaderSize() (uint32, error) {
	dec.buf.Reset()
	if _, err := io.CopyN(dec.buf, dec.r, 4); err != nil {
		return 0, err
	}

	size := binary.BigEndian.Uint32(dec.buf.Bytes())

	if size >= maxBlobHeaderSize {
		return 0, errors.New("BlobHeader size >= 64Kb")
	}
	return size, nil
}

func (dec *Decoder) readBlobHeader(size uint32) (*OSMPBF.BlobHeader, error) {
	dec.buf.Reset()
	if _, err := io.CopyN(dec.buf, dec.r, int64(size)); err != nil {
		return nil, err
	}

	blobHeader := new(OSMPBF.BlobHeader)
	if err := proto.Unmarshal(dec.buf.Bytes(), blobHeader); err != nil {
		return nil, err
	}

	if blobHeader.GetDatasize() >= MaxBlobSize {
		return nil, errors.New("Blob size >= 32Mb")
	}
	return blobHeader, nil
}

func (dec *Decoder) readBlob(blobHeader *OSMPBF.BlobHeader) (*OSMPBF.Blob, error) {
	dec.buf.Reset()
	if _, err := io.CopyN(dec.buf, dec.r, int64(blobHeader.GetDatasize())); err != nil {
		return nil, err
	}

	blob := new(OSMPBF.Blob)
	if err := proto.Unmarshal(dec.buf.Bytes(), blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func getData(blob *OSMPBF.Blob) ([]byte, error) {
	switch {
	case blob.Raw != nil:
		return blob.GetRaw(), nil

	case blob.ZlibData != nil:
		r, err := zlib.NewReader(bytes.NewReader(blob.GetZlibData()))
		if err != nil {
			return nil, err
		}
		buf := bytes.NewBuffer(make([]byte, 0, blob.GetRawSize()+bytes.MinRead))
		_, err = buf.ReadFrom(r)
		if err != nil {
			return nil, err
		}
		if buf.Len() != int(blob.GetRawSize()) {
			err = fmt.Errorf("raw blob data size %d but expected %d", buf.Len(), blob.GetRawSize())
			return nil, err
		}
		return buf.Bytes(), nil

	default:
		return nil, errors.New("unknown blob data")
	}
}

func (dec *Decoder) readOSMHeader() error {
	var err error
	dec.headerOnce.Do(func() {
		var blobHeader *OSMPBF.BlobHeader
		var blob *OSMPBF.Blob
		blobHeader, blob, err = dec.readFileBlock()
		if err == nil {
			if blobHeader.GetType() == "OSMHeader" {
				err = dec.decodeOSMHeader(blob)
			} else {
				err = fmt.Errorf("unexpected first fileblock of type %s", blobHeader.GetType())
			}
		}
	})

	return err
}

func (dec *Decoder) decodeOSMHeader(blob *OSMPBF.Blob) error {
	data, err := getData(blob)
	if err != nil {
		return err
	}

	headerBlock := new(OSMPBF.HeaderBlock)
	if err := proto.Unmarshal(data, headerBlock); err != nil {
		return err
	}

	// Check we have the parse capabilities
	requiredFeatures := headerBlock.GetRequiredFeatures()
	for _, feature := range requiredFeatures {
		if !parseCapabilities[feature] {
			return fmt.Errorf("parser does not have %s capability", feature)
		}
	}

	// Read properties to header struct
	header := &Header{
		RequiredFeatures: headerBlock.GetRequiredFeatures(),
		OptionalFeatures: headerBlock.GetOptionalFeatures(),
		WritingProgram:   headerBlock.GetWritingprogram(),
		Source:           headerBlock.GetSource(),
		OsmosisReplicationBaseUrl:        headerBlock.GetOsmosisReplicationBaseUrl(),
		OsmosisReplicationSequenceNumber: headerBlock.GetOsmosisReplicationSequenceNumber(),
	}

	// convert timestamp epoch seconds to golang time structure if it exists
	if headerBlock.OsmosisReplicationTimestamp != nil {
		header.OsmosisReplicationTimestamp = time.Unix(*headerBlock.OsmosisReplicationTimestamp, 0)
	}
	// read bounding box if it exists
	if headerBlock.Bbox != nil {
		// Units are always in nanodegree and do not obey granularity rules. See osmformat.proto
		header.BoundingBox = &BoundingBox{
			Left:   1e-9 * float64(*headerBlock.Bbox.Left),
			Right:  1e-9 * float64(*headerBlock.Bbox.Right),
			Bottom: 1e-9 * float64(*headerBlock.Bbox.Bottom),
			Top:    1e-9 * float64(*headerBlock.Bbox.Top),
		}
	}

	dec.header = header

	return nil
}
