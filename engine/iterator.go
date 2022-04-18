package engine

// ListIterator is an iterator for a proper list.
type ListIterator struct {
	List Term
	Env  *Env

	current, rest Term
	err           error
	visited       map[*Compound]struct{}
}

// Next proceeds to the next element of the list and returns true if there's such an element.
func (i *ListIterator) Next() bool {
	if i.rest == nil {
		i.rest = i.List
	}
	if i.visited == nil {
		i.visited = map[*Compound]struct{}{}
	}

	switch l := i.Env.Resolve(i.rest).(type) {
	case Variable:
		i.err = ErrInstantiation
		return false
	case Atom:
		if l != "[]" {
			i.err = TypeErrorList(i.List, i.Env)
		}
		return false
	case *Compound:
		if l.Functor != "." || len(l.Args) != 2 {
			i.err = TypeErrorList(i.List, i.Env)
			return false
		}

		if _, ok := i.visited[l]; ok {
			i.err = TypeErrorList(i.List, i.Env)
			return false
		}
		i.visited[l] = struct{}{}

		i.current, i.rest = l.Args[0], l.Args[1]
		return true
	default:
		i.err = TypeErrorList(i.List, i.Env)
		return false
	}
}

// Current returns the current element.
func (i *ListIterator) Current() Term {
	return i.current
}

// Err returns an error.
func (i *ListIterator) Err() error {
	return i.err
}

// SeqIterator is an iterator for a sequence.
type SeqIterator struct {
	Seq Term
	Env *Env

	current Term
}

// Next proceeds to the next element of the sequence and returns true if there's such an element.
func (i *SeqIterator) Next() bool {
	switch s := i.Env.Resolve(i.Seq).(type) {
	case nil:
		return false
	case *Compound:
		if s.Functor != "," || len(s.Args) != 2 {
			i.current = s
			i.Seq = nil
			return true
		}
		i.Seq = s.Args[1]
		i.current = s.Args[0]
		return true
	default:
		i.current = s
		i.Seq = nil
		return true
	}
}

// Current returns the current element.
func (i *SeqIterator) Current() Term {
	return i.current
}

// AltIterator is an iterator for alternatives.
type AltIterator struct {
	Alt Term
	Env *Env

	current Term
}

// Next proceeds to the next element of the alternatives and returns true if there's such an element.
func (i *AltIterator) Next() bool {
	switch a := i.Env.Resolve(i.Alt).(type) {
	case nil:
		return false
	case *Compound:
		if a.Functor != ";" || len(a.Args) != 2 {
			i.current = a
			i.Alt = nil
			return true
		}

		// if-then-else construct
		if c, ok := i.Env.Resolve(a.Args[0]).(*Compound); ok && c.Functor == "->" && len(c.Args) == 2 {
			i.current = a
			i.Alt = nil
			return true
		}

		i.Alt = a.Args[1]
		i.current = a.Args[0]
		return true
	default:
		i.current = a
		i.Alt = nil
		return true
	}
}

// Current returns the current element.
func (i *AltIterator) Current() Term {
	return i.current
}

// AnyIterator is an iterator for a list or a sequence.
type AnyIterator struct {
	Any Term
	Env *Env

	backend interface {
		Next() bool
		Current() Term
	}
}

// Next proceeds to the next element and returns true if there's such an element.
func (i *AnyIterator) Next() bool {
	if i.backend == nil {
		if a, ok := i.Env.Resolve(i.Any).(*Compound); ok && a.Functor == "." && len(a.Args) == 2 {
			i.backend = &ListIterator{List: i.Any, Env: i.Env}
		} else {
			i.backend = &SeqIterator{Seq: i.Any, Env: i.Env}
		}
	}

	return i.backend.Next()
}

// Current returns the current element.
func (i *AnyIterator) Current() Term {
	return i.backend.Current()
}

// Err returns an error.
func (i *AnyIterator) Err() error {
	b, ok := i.backend.(interface{ Err() error })
	if !ok {
		return nil
	}
	return b.Err()
}
