// Package ber provides functions for parsing BER (Basic Encoding Rules) encoded data.
package ber

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/ubavic/bas-celik/v2/card/carderrors"
)

// BER represents a node (or a tree) of a BER structure.
// Each leaf node contains data, and it is considered 'primitive'.
// Non-leaf nodes don't contain any data, but they contain references to child nodes.
type BER struct {
	tag       uint32 // Complete tag of a node.
	primitive bool   // Denotes if node is a leaf.
	data      []byte // Data of leaf node. Should only exist if primitive is true.
	children  []BER  // Branch nodes children. Should only exist if primitive is false.
}

// ParseBER parses BER data (described in ISO/IEC 7816-4 (2005)).
func ParseBER(data []byte) (*BER, error) {
	primitive, constructed, err := parseBERLayer(data)

	if err != nil {
		return nil, err
	}

	ber := BER{
		tag:       0,
		primitive: false,
		data:      nil,
		children:  []BER{},
	}

	for t, v := range primitive {
		val := BER{
			tag:       t,
			primitive: true,
			data:      v,
			children:  nil,
		}

		err = ber.add(val)
		if err != nil {
			return nil, fmt.Errorf("adding primitive value: %w", err)
		}
	}

	for t, v := range constructed {
		subBer, err := ParseBER(v)
		if err != nil {
			return nil, err
		}
		val := BER{
			tag:       t,
			primitive: false,
			data:      nil,
			children:  subBer.children,
		}

		err = ber.add(val)
		if err != nil {
			return nil, fmt.Errorf("adding primitive value: %w", err)
		}
	}

	return &ber, nil
}

// Access node's data with the provided address composed as a list of tags.
func (tree BER) access(address ...uint32) ([]byte, error) {
	if len(address) == 0 {
		return tree.data, nil
	}
	var found *BER
	for i := range tree.children {
		if tree.children[i].tag == address[0] {
			found = &tree.children[i]
			break
		}
	}
	if found != nil {
		return found.access(address[1:]...)
	}
	return nil, errors.New("tag not found")
}

// Recursively inserts a new node (with all children nodes) into BER tree. It doesn't copy data.
// If a node with the same tag and the type (primitive/constructed) already exists in tree, then procedure continues
// inserting in deeper levels. If a node with the same tag and different type already exists, function return error.
func (tree *BER) add(newBER BER) error {
	if tree.primitive {
		return errors.New("can't add a value into primitive value")
	}

	var targetField *BER

	alreadyExists := false
	for i := range tree.children {
		if tree.children[i].tag == newBER.tag {
			alreadyExists = true
			targetField = &tree.children[i]
			break
		}
	}

	if !alreadyExists {
		tree.children = append(tree.children, newBER)
		return nil
	}
	if targetField.primitive == newBER.primitive {
		if targetField.primitive {
			*targetField = newBER
		} else {
			for i := range newBER.children {
				err := targetField.add(newBER.children[i])
				if err != nil {
					return err
				}
			}
		}
	} else {
		return errors.New("types don't match")
	}

	return nil
}

// Merge merges two BER trees by adding all nodes of the second tree into the first tree.
// Doesn't copy any data.
func (tree *BER) Merge(newBER BER) error {
	if tree.tag != newBER.tag {
		return errors.New("tags don't match")
	}

	for _, c := range newBER.children {
		if err := tree.add(c); err != nil {
			return err
		}
	}

	return nil
}

// Parses one level of BER-TLV encoded data.
// Returns map of primitive and constructed fields.
func parseBERLayer(data []byte) (map[uint32][]byte, map[uint32][]byte, error) {
	primF := make(map[uint32][]byte)
	consF := make(map[uint32][]byte)
	offset := uint32(0)

	for {
		tag, primitive, offsetDelta, err := ParseTag(data[offset:])
		if err != nil {
			return nil, nil, fmt.Errorf("parsing BER record tag: %w", err)
		}

		dataLenInt := len(data)
		if dataLenInt > math.MaxUint32 {
			return nil, nil, carderrors.ErrInvalidLength
		}
		dataLength := uint32(dataLenInt)

		offset += offsetDelta

		length, offsetDelta, err := ParseLength(data[offset:])
		if err != nil {
			return nil, nil, fmt.Errorf("parsing BER record length: %w", err)
		}

		if dataLength <= offset+length {
			return nil, nil, fmt.Errorf("parsing BER record data: %w", carderrors.ErrInvalidLength)
		}

		offset += offsetDelta
		value := data[offset : offset+length]

		if primitive {
			primF[tag] = value
		} else {
			consF[tag] = value
		}

		offset += length

		if offset == dataLength {
			break
		} else if offset > dataLength {
			return nil, nil, carderrors.ErrInvalidLength
		}
	}

	return primF, consF, nil
}

// AssignFrom assigns a string value from the BER tree at the specified address path.
func (tree *BER) AssignFrom(target *string, address ...uint32) {
	bytes, err := tree.access(address...)
	if err == nil {
		*target = string(bytes)
	}
}

// Flattens a BER tree into list of strings. Used for printing.
func (tree *BER) levels() []string {
	if tree.primitive {
		return []string{fmt.Sprintf("%X: %s", tree.tag, string(tree.data))}
	}
	strings := []string{fmt.Sprint(tree.tag) + ":"}
	for _, child := range tree.children {
		childrenStrings := child.levels()
		for i := range childrenStrings {
			childrenStrings[i] = "  " + childrenStrings[i]
		}
		strings = append(strings, childrenStrings...)
	}

	return strings
}

// String flattens a BER tree into a single string. Each line represents a single node of a tree.
func (tree BER) String() string {
	return strings.Join(tree.levels(), "\n")
}

// ParseLength parses the length of a field according to specification given in ISO 7816-4 (5. Organization for interchange).
// Returns parsed length, number of parsed bytes and possible error.
func ParseLength(data []byte) (uint32, uint32, error) {
	if len(data) == 0 {
		return 0, 0, carderrors.ErrInvalidLength
	}

	firstByte := uint32(data[0])
	var offset, length uint32
	if firstByte < 0x80 {
		length = uint32(data[0])
		offset = 1
	} else if firstByte == 0x80 {
		return 0, 0, carderrors.ErrInvalidFormat
	} else if firstByte == 0x81 && len(data) >= 2 {
		length = uint32(data[1])
		offset = 2
	} else if firstByte == 0x82 && len(data) >= 3 {
		length = uint32(binary.BigEndian.Uint16(data[1:]))
		offset = 3
	} else if firstByte == 0x83 && len(data) >= 4 {
		length = 0x00FFFFFF & binary.BigEndian.Uint32(data)
		offset = 4
	} else if firstByte == 0x84 && len(data) >= 5 {
		length = binary.BigEndian.Uint32(data[1:])
		offset = 5
	} else {
		return 0, 0, carderrors.ErrInvalidLength
	}

	return length, offset, nil
}

// ParseTag parses the tag of a field according to specification given in ISO 7816-4 (5. Organization for interchange).
// Returns parsed tag, primitive flag, number of parsed bytes and possible error.
func ParseTag(data []byte) (uint32, bool, uint32, error) {
	if len(data) == 0 {
		return 0, false, 0, carderrors.ErrInvalidLength
	}

	primitive := data[0]&0b100000 == 0

	var tag, offset uint32
	if 0x1F&data[0] != 0x1F {
		tag = uint32(data[0])
		offset = 1
	} else if len(data) >= 2 && data[1]&0x80 == 0x00 {
		tag = uint32(binary.BigEndian.Uint16(data))
		offset = 2
	} else if len(data) >= 3 {
		tag = uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
		offset = 3
	} else {
		return 0, false, 0, carderrors.ErrInvalidLength
	}

	return tag, primitive, offset, nil
}
