/*
	Inspired in:

	https://github.com/mimblewimble/grin/blob/845c41de13e9bdeb0f0b4667fbc7ef8be921a2f4/core/src/core/pmmr/pmmr.rs
	https://github.com/paritytech/merkle-mountain-range/blob/8a8a2dd5d172545faac314f3e7b6a43a85395c03/src/mmr.rs

*/

package mmr

import (
	"errors"
	"math/bits"
)

var (
	errorInconsistentStore = errors.New("inconsistent store")
	errorGetRootOnEmpty    = errors.New("get root on empty MMR")
)

type MMRMergeFunction func(left, right MMRElement) (MMRElement, error)

type MMRElement []byte

type MMRNode struct {
	pos      uint64
	elements []MMRElement
}

type MMR struct {
	size  uint64
	batch *MMRBatch
	merge MMRMergeFunction
}

func NewMMR(size uint64, batch *MMRBatch, merge MMRMergeFunction) *MMR {
	return &MMR{
		size:  size,
		batch: batch,
		merge: merge,
	}
}

func (mmr *MMR) Root() (MMRElement, error) {
	if mmr.size == 0 {
		return nil, errorGetRootOnEmpty
	} else if mmr.size == 1 {
		root, err := mmr.batch.getElement(0)
		if err != nil || root == nil {
			return nil, errorInconsistentStore
		}
		return *root, nil
	}

	peaksPosition := mmr.getPeaks()
	peaks := make([]MMRElement,0)

	for _, pos := range peaksPosition {
		peak, err := mmr.batch.getElement(pos)
		if err != nil || peak == nil {
			return nil, errorInconsistentStore
		}
		peaks = append(peaks, *peak)
	}

	return mmr.bagPeaks(peaks)
}

func (mmr *MMR) Push(leaf MMRElement) (uint64, error) {
	elements := []MMRElement{leaf}
	peakMap := mmr.peakMap()
	elemPosition := mmr.size
	position := mmr.size
	peak := uint64(1)

	for (peakMap & peak) != 0 {
		peak <<= 1
		position += 1
		leftPosition := uint64(position - peak)
		leftElement, err := mmr.findElement(leftPosition, elements)

		if err != nil {
			return 0, err
		}

		rightElement := elements[len(elements)-1] // TODO: check this wont fail
		parentElement, err := mmr.merge(leftElement, rightElement)

		if err != nil {
			return 0, err
		}

		elements = append(elements, parentElement)
	}

	mmr.batch.append(elemPosition, elements)
	mmr.size = position + 1
	return position, nil
}

func (mmr *MMR) findElement(position uint64, values []MMRElement) (MMRElement, error) {
	if position > mmr.size {
		positionOffset := position - mmr.size
		return values[positionOffset], nil
	}
	
	value, err := mmr.batch.getElement(position)
	if err != nil || value == nil {
		return nil, errorInconsistentStore
	}

	return *value, nil
}

func (mmr *MMR) peakMap() uint64 {
	if mmr.size == 0 {
		return 0
	}

	pos := mmr.size
	peakSize := ^uint64(0) >> bits.LeadingZeros64(pos)
	peakMap := uint64(0)

	for peakSize > 0 {
		peakMap <<= 1
		if pos >= peakSize {
			pos -= peakSize
			peakMap |= 1
		}
		peakSize >>= 1
	}

	return peakMap
}

func (mmr *MMR) getPeaks() []uint64 {
	if mmr.size == 0 {
		return []uint64{}
	}

	pos := mmr.size
	peakSize := ^uint64(0) >> bits.LeadingZeros64(pos)
	peaks := make([]uint64, 0)
	peaksSum := uint64(0)
	for peakSize > 0 {
		if pos >= peakSize {
			pos -= peakSize
			peaks = append(peaks, uint64(peaksSum) + peakSize - 1)
			peaksSum += peakSize
		}
		peakSize >>= 1
	}

	return peaks
}

func (mmr *MMR) bagPeaks(peaks []MMRElement) (MMRElement, error) {
	for len(peaks) > 1 {
		rightPeak, peaks := peaks[len(peaks)-1], peaks[:len(peaks)-1]
		leftPeak, peaks := peaks[len(peaks)-1], peaks[:len(peaks)-1]

		mergedPeak, err := mmr.merge(leftPeak, rightPeak)
		if err != nil {
			return nil, err
		}
		peaks = append(peaks, mergedPeak)
	}

	return peaks[0], nil
}