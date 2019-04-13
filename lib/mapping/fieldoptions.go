package mapping

type FieldOptions struct {
	FromString bool
	Optional   bool
	Options    []string
	Default    string
}

func (o *FieldOptions) getDefault() (string, bool) {
	if o == nil {
		return "", false
	} else {
		return o.Default, len(o.Default) > 0
	}
}

func (o *FieldOptions) fromString() bool {
	return o != nil && o.FromString
}

func (o *FieldOptions) optional() bool {
	return o != nil && o.Optional
}

func (o *FieldOptions) options() []string {
	if o == nil {
		return nil
	}

	return o.Options
}
