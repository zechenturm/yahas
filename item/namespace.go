package item

type Namespace map[string]*Item

type NotFoundError struct {
	itemName string
}

func (nfe NotFoundError) Error() string{
	return "item " + nfe.itemName +  "not found"
}

func (ns *Namespace) Get(name string) (*Item, error){
	itm, ok := (*ns)[name]
	if !ok {
		return nil, NotFoundError{name}
	}
	return itm, nil

}