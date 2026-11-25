package text

import (
	"unicode"
	"unicode/utf8"

	"github.com/go-text/typesetting/bidi"
)

// fallbackBytes is how long a chunk may grow without a sentence end before
// the next space or tab cuts it.
const fallbackBytes = 4 * 1024

// firstRTLRune is the first code point that can be strong right-to-left or
// an explicit bidi control (Hebrew starts at U+0590). A line of smaller runes
// is one level-0 run and skips the bidi pass.
const firstRTLRune = 0x0590

// appendCuts appends to cuts the end offset of every chunk of line, in order;
// the last is len(line). The rule is the one Ebitengine's text/v2 shapes by,
// which this package matches on purpose:
//   - a run whose bidi level is odd (right to left under our left-to-right
//     base) is one chunk, since cutting it would let its edges re-resolve;
//   - elsewhere a chunk ends after a sentence terminator cluster and the
//     marks that follow it, if a terminator of the "STerm" kind, or if only
//     "." kinds follow with white space and then no lower-case letter;
//   - after fallbackBytes without such an end, a chunk ends after the next
//     space or tab and the marks that follow it.
//
// line holds no line break.
func appendCuts(cuts []int, line string, paragraph *bidi.Paragraph) []int {
	if line == "" {
		return cuts
	}
	if !mayBeBidi(line) {
		return appendRunCuts(cuts, line, 0)
	}
	runs := paragraph.SegmentString(line, bidi.LeftToRight)
	if runs.NumRuns() == 0 {
		return append(cuts, len(line))
	}
	var byteIndex, runeIndex int
	advanceTo := func(runeEnd int) int {
		for runeIndex < runeEnd {
			_, width := utf8.DecodeRuneInString(line[byteIndex:])
			byteIndex += width
			runeIndex++
		}
		return byteIndex
	}
	for i := range runs.NumRuns() {
		run := runs.Run(i)
		start := advanceTo(run.Start)
		end := advanceTo(run.End)
		if run.Level%2 == 1 {
			cuts = append(cuts, end)
			continue
		}
		cuts = appendRunCuts(cuts, line[start:end], start)
	}
	return cuts
}

func mayBeBidi(line string) bool {
	for _, char := range line {
		if char >= firstRTLRune {
			return true
		}
	}
	return false
}

// appendRunCuts cuts one run of a single even level; offset puts the cuts in
// the coordinates of the whole line.
func appendRunCuts(cuts []int, run string, offset int) []int {
	chunkStart := 0
	for index := 0; index < len(run); {
		char, width := utf8.DecodeRuneInString(run[index:])
		if (char == ' ' || char == '\t') && index-chunkStart >= fallbackBytes {
			index = skipMarks(run, index+width)
			cuts = append(cuts, offset+index)
			chunkStart = index
			continue
		}
		if !isTerminator(char) {
			index += width
			continue
		}
		clusterEnd, hasSentenceEnd := terminatorCluster(run, index)
		if hasSentenceEnd || endsSentenceAfterSpace(run, clusterEnd) {
			clusterEnd = skipMarks(run, clusterEnd)
			cuts = append(cuts, offset+clusterEnd)
			chunkStart = clusterEnd
		}
		index = clusterEnd
	}
	return append(cuts, offset+len(run))
}

// terminatorCluster is the end of the run of terminators that starts at
// start, and whether one of them ends a sentence by itself ("!", "?", not ".").
func terminatorCluster(run string, start int) (end int, hasSentenceEnd bool) {
	end = start
	for end < len(run) {
		char, width := utf8.DecodeRuneInString(run[end:])
		if !isTerminator(char) {
			break
		}
		hasSentenceEnd = hasSentenceEnd || !isFullStop(char)
		end += width
	}
	return end, hasSentenceEnd
}

// endsSentenceAfterSpace is true when white space follows a cluster of full
// stops and the next visible character is not lower case ("Hi. There" cuts,
// "etc. and so" does not).
func endsSentenceAfterSpace(run string, clusterEnd int) bool {
	index := clusterEnd
	for index < len(run) {
		char, width := utf8.DecodeRuneInString(run[index:])
		if !unicode.IsSpace(char) {
			break
		}
		index += width
	}
	if index == clusterEnd {
		return false
	}
	if index == len(run) {
		return true
	}
	next, _ := utf8.DecodeRuneInString(run[index:])
	return !unicode.IsLower(next)
}

// skipMarks moves past combining marks and format characters that follow
// position, so a cut does not strand them from their base.
func skipMarks(run string, position int) int {
	for position < len(run) {
		char, width := utf8.DecodeRuneInString(run[position:])
		if char < utf8.RuneSelf || !(unicode.IsMark(char) || unicode.Is(unicode.Cf, char)) {
			return position
		}
		position += width
	}
	return position
}

// isFullStop is the sentence terminators that do not end a sentence alone
// (Unicode Sentence_Break ATerm).
func isFullStop(char rune) bool {
	switch char {
	case '.', 0x2024, 0xFE52, 0xFF0E:
		return true
	}
	return false
}

// isTerminator is every sentence terminator: ATerm and STerm. Sentence_Break
// STerm is the Sentence_Terminal property without the ATerm characters, plus
// the two characters Unicode 16 added that Go's tables may not have yet (a
// check against Ebitengine's generated table found only these two missing).
func isTerminator(char rune) bool {
	return unicode.Is(unicode.Sentence_Terminal, char) || char == 0x2E60 || char == 0x2E61
}
