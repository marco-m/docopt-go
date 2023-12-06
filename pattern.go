package docopt

import (
	"fmt"
	"reflect"
	"strings"
)

type patternType uint

const (
	// leaf
	patternArgument patternType = 1 << iota
	patternCommand
	patternOption

	// branch
	patternRequired
	patternOptionAL
	patternOptionSSHORTCUT // Marker/placeholder for [options] shortcut.
	patternOneOrMore
	patternEither

	patternLeaf = patternArgument +
		patternCommand +
		patternOption
	patternBranch = patternRequired +
		patternOptionAL +
		patternOptionSSHORTCUT +
		patternOneOrMore +
		patternEither
	patternAll     = patternLeaf + patternBranch
	patternDefault = 0
)

func (pt patternType) String() string {
	switch pt {
	case patternArgument:
		return "argument"
	case patternCommand:
		return "command"
	case patternOption:
		return "option"
	case patternRequired:
		return "required"
	case patternOptionAL:
		return "optional"
	case patternOptionSSHORTCUT:
		return "optionsshortcut"
	case patternOneOrMore:
		return "oneormore"
	case patternEither:
		return "either"
	case patternLeaf:
		return "leaf"
	case patternBranch:
		return "branch"
	case patternAll:
		return "all"
	case patternDefault:
		return "default"
	}
	return ""
}

type pattern struct {
	Type patternType

	Children patternList

	Name  string
	Value any

	Short    string
	Long     string
	ArgCount int
}

type patternList []*pattern

func newBranchPattern(pt patternType, pl ...*pattern) *pattern {
	var pat pattern
	pat.Type = pt
	pat.Children = make(patternList, len(pl))
	copy(pat.Children, pl)
	return &pat
}

func newRequired(pl ...*pattern) *pattern {
	return newBranchPattern(patternRequired, pl...)
}

func newEither(pl ...*pattern) *pattern {
	return newBranchPattern(patternEither, pl...)
}

func newOneOrMore(pl ...*pattern) *pattern {
	return newBranchPattern(patternOneOrMore, pl...)
}

func newOptional(pl ...*pattern) *pattern {
	return newBranchPattern(patternOptionAL, pl...)
}

func newOptionsShortcut() *pattern {
	var pat pattern
	pat.Type = patternOptionSSHORTCUT
	return &pat
}

func newLeafPattern(pt patternType, name string, value any) *pattern {
	// default: value=nil
	var pat pattern
	pat.Type = pt
	pat.Name = name
	pat.Value = value
	return &pat
}

func newArgument(name string, value any) *pattern {
	// default: value=nil
	return newLeafPattern(patternArgument, name, value)
}

func newCommand(name string, value any) *pattern {
	// default: value=false
	var pat pattern
	pat.Type = patternCommand
	pat.Name = name
	pat.Value = value
	return &pat
}

func newOption(short, long string, argCount int, value any) *pattern {
	// default: "", "", 0, false
	var pat pattern
	pat.Type = patternOption
	pat.Short = short
	pat.Long = long
	if long != "" {
		pat.Name = long
	} else {
		pat.Name = short
	}
	pat.ArgCount = argCount
	if value == false && argCount > 0 {
		pat.Value = nil
	} else {
		pat.Value = value
	}
	return &pat
}

func (pat *pattern) flat(pt patternType) (patternList, error) {
	if pat.Type&patternLeaf != 0 {
		if pt == patternDefault {
			pt = patternAll
		}
		if pat.Type&pt != 0 {
			return patternList{pat}, nil
		}
		return patternList{}, nil
	}

	if pat.Type&patternBranch != 0 {
		if pat.Type&pt != 0 {
			return patternList{pat}, nil
		}
		result := patternList{}
		for _, child := range pat.Children {
			childFlat, err := child.flat(pt)
			if err != nil {
				return nil, err
			}
			result = append(result, childFlat...)
		}
		return result, nil
	}
	return nil, fmt.Errorf("unknown pattern type: %d, %d", pat.Type, pt)
}

func (pat *pattern) fix() error {
	err := pat.fixIdentities(nil)
	if err != nil {
		return err
	}
	pat.fixRepeatingArguments()
	return nil
}

func (pat *pattern) fixIdentities(uniq patternList) error {
	// Make pattern-tree tips point to same object if they are equal.
	if pat.Type&patternBranch == 0 {
		return nil
	}
	if uniq == nil {
		pFlat, err := pat.flat(patternDefault)
		if err != nil {
			return err
		}
		uniq = pFlat.unique()
	}
	for i, child := range pat.Children {
		if child.Type&patternBranch == 0 {
			ind, err := uniq.index(child)
			if err != nil {
				return err
			}
			pat.Children[i] = uniq[ind]
		} else {
			err := child.fixIdentities(uniq)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (pat *pattern) fixRepeatingArguments() {
	// Fix elements that should accumulate/increment values.
	var either []patternList

	for _, child := range pat.transform().Children {
		either = append(either, child.Children)
	}
	for _, cas := range either {
		casMultiple := patternList{}
		for _, e := range cas {
			if cas.count(e) > 1 {
				casMultiple = append(casMultiple, e)
			}
		}
		for _, e := range casMultiple {
			if e.Type == patternArgument || e.Type == patternOption && e.ArgCount > 0 {
				switch v := e.Value.(type) {
				case string:
					e.Value = strings.Fields(v)
				case []string:
				default:
					e.Value = []string{}
				}
			}
			if e.Type == patternCommand || e.Type == patternOption && e.ArgCount == 0 {
				e.Value = 0
			}
		}
	}
}

func (pat *pattern) match(left *patternList, collected *patternList) (bool, *patternList, *patternList) {
	if collected == nil {
		collected = &patternList{}
	}
	if pat.Type&patternRequired != 0 {
		l := left
		c := collected
		for _, p := range pat.Children {
			var matched bool
			matched, l, c = p.match(l, c)
			if !matched {
				return false, left, collected
			}
		}
		return true, l, c
	} else if pat.Type&patternOptionAL != 0 || pat.Type&patternOptionSSHORTCUT != 0 {
		for _, p := range pat.Children {
			_, left, collected = p.match(left, collected)
		}
		return true, left, collected
	} else if pat.Type&patternOneOrMore != 0 {
		if len(pat.Children) != 1 {
			panic("OneOrMore.match(): assert len(pat.Children) == 1")
		}
		l := left
		c := collected
		var lAlt *patternList
		matched := true
		times := 0
		for matched {
			// could it be that something didn't match but changed l or c?
			matched, l, c = pat.Children[0].match(l, c)
			if matched {
				times++
			}
			if lAlt == l {
				break
			}
			lAlt = l
		}
		if times >= 1 {
			return true, l, c
		}
		return false, left, collected
	} else if pat.Type&patternEither != 0 {
		type outcomeStruct struct {
			matched   bool
			left      *patternList
			collected *patternList
			length    int
		}
		outcomes := []outcomeStruct{}
		for _, p := range pat.Children {
			matched, l, c := p.match(left, collected)
			outcome := outcomeStruct{matched, l, c, len(*l)}
			if matched {
				outcomes = append(outcomes, outcome)
			}
		}
		if len(outcomes) > 0 {
			minLen := outcomes[0].length
			minIndex := 0
			for i, v := range outcomes {
				if v.length < minLen {
					minIndex = i
				}
			}
			return outcomes[minIndex].matched, outcomes[minIndex].left, outcomes[minIndex].collected
		}
		return false, left, collected
	} else if pat.Type&patternLeaf != 0 {
		pos, match := pat.singleMatch(left)
		var increment any
		if match == nil {
			return false, left, collected
		}
		leftAlt := make(patternList, len((*left)[:pos]), len((*left)[:pos])+len((*left)[pos+1:]))
		copy(leftAlt, (*left)[:pos])
		leftAlt = append(leftAlt, (*left)[pos+1:]...)
		sameName := patternList{}
		for _, a := range *collected {
			if a.Name == pat.Name {
				sameName = append(sameName, a)
			}
		}

		switch pat.Value.(type) {
		case int, []string:
			switch pat.Value.(type) {
			case int:
				increment = 1
			case []string:
				switch v := match.Value.(type) {
				case string:
					increment = []string{v}
				default:
					increment = match.Value
				}
			}
			if len(sameName) == 0 {
				match.Value = increment
				collectedMatch := make(patternList, len(*collected), len(*collected)+1)
				copy(collectedMatch, *collected)
				collectedMatch = append(collectedMatch, match)
				return true, &leftAlt, &collectedMatch
			}
			switch sameName[0].Value.(type) {
			case int:
				sameName[0].Value = sameName[0].Value.(int) + increment.(int)
			case []string:
				sameName[0].Value = append(sameName[0].Value.([]string), increment.([]string)...)
			}
			return true, &leftAlt, collected
		}
		collectedMatch := make(patternList, len(*collected), len(*collected)+1)
		copy(collectedMatch, *collected)
		collectedMatch = append(collectedMatch, match)
		return true, &leftAlt, &collectedMatch
	}
	panic("unmatched type")
}

func (pat *pattern) singleMatch(left *patternList) (int, *pattern) {
	if pat.Type&patternArgument != 0 {
		for n, p2 := range *left {
			if p2.Type&patternArgument != 0 {
				return n, newArgument(pat.Name, p2.Value)
			}
		}
		return -1, nil
	} else if pat.Type&patternCommand != 0 {
		for n, p2 := range *left {
			if p2.Type&patternArgument != 0 {
				if p2.Value == pat.Name {
					return n, newCommand(pat.Name, true)
				}
				break
			}
		}
		return -1, nil
	} else if pat.Type&patternOption != 0 {
		for n, p2 := range *left {
			if pat.Name == p2.Name {
				return n, p2
			}
		}
		return -1, nil
	}
	panic("unmatched type")
}

func (pat *pattern) String() string {
	if pat.Type&patternOption != 0 {
		return fmt.Sprintf("%s(%s, %s, %d, %+v)", pat.Type, pat.Short, pat.Long, pat.ArgCount, pat.Value)
	} else if pat.Type&patternLeaf != 0 {
		return fmt.Sprintf("%s(%s, %+v)", pat.Type, pat.Name, pat.Value)
	} else if pat.Type&patternBranch != 0 {
		result := ""
		for i, child := range pat.Children {
			if i > 0 {
				result += ", "
			}
			result += child.String()
		}
		return fmt.Sprintf("%s(%s)", pat.Type, result)
	}
	panic("unmatched type")
}

func (pat *pattern) transform() *pattern {
	/*
		Expand pattern into an (almost) equivalent one, but with single Either.

		Example: ((-a | -b) (-c | -d)) => (-a -c | -a -d | -b -c | -b -d)
		Quirks: [-a] => (-a), (-a...) => (-a -a)
	*/
	result := []patternList{}
	groups := []patternList{{pat}}
	parents := patternRequired +
		patternOptionAL +
		patternOptionSSHORTCUT +
		patternEither +
		patternOneOrMore
	for len(groups) > 0 {
		children := groups[0]
		groups = groups[1:]
		var child *pattern
		for _, c := range children {
			if c.Type&parents != 0 {
				child = c
				break
			}
		}
		if child != nil {
			children.remove(child)
			if child.Type&patternEither != 0 {
				for _, c := range child.Children {
					pl := patternList{}
					pl = append(pl, c)
					pl = append(pl, children...)
					groups = append(groups, pl)
				}
			} else if child.Type&patternOneOrMore != 0 {
				pl := patternList{}
				pl = append(pl, child.Children.double()...)
				pl = append(pl, children...)
				groups = append(groups, pl)
			} else {
				pl := patternList{}
				pl = append(pl, child.Children...)
				pl = append(pl, children...)
				groups = append(groups, pl)
			}
		} else {
			result = append(result, children)
		}
	}
	either := patternList{}
	for _, e := range result {
		either = append(either, newRequired(e...))
	}
	return newEither(either...)
}

func (pat *pattern) eq(other *pattern) bool {
	return reflect.DeepEqual(pat, other)
}

func (pl patternList) unique() patternList {
	table := make(map[string]bool)
	result := patternList{}
	for _, p2 := range pl {
		if !table[p2.String()] {
			table[p2.String()] = true
			result = append(result, p2)
		}
	}
	return result
}

func (pl patternList) index(p *pattern) (int, error) {
	for i, p2 := range pl {
		if p2.eq(p) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("%s not in list", p)
}

func (pl patternList) count(p1 *pattern) int {
	count := 0
	for _, p2 := range pl {
		if p2.eq(p1) {
			count++
		}
	}
	return count
}

func (pl patternList) diff(pl2 patternList) patternList {
	lAlt := make(patternList, len(pl2))
	copy(lAlt, pl2)
	result := make(patternList, 0, len(pl))
	for _, p1 := range pl {
		if p1 == nil {
			continue
		}
		match := false
		for i, w := range lAlt {
			if w.eq(p1) {
				match = true
				lAlt[i] = nil
				break
			}
		}
		if !match {
			result = append(result, p1)
		}
	}
	return result
}

func (pl patternList) double() patternList {
	l := len(pl)
	result := make(patternList, l*2)
	copy(result, pl)
	copy(result[l:2*l], pl)
	return result
}

func (pl *patternList) remove(p *pattern) {
	*pl = pl.diff(patternList{p})
}

func (pl patternList) dictionary() map[string]any {
	dict := make(map[string]any)
	for _, p1 := range pl {
		dict[p1.Name] = p1.Value
	}
	return dict
}
