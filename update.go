package main

type UpdateEvent interface {
	Category() string
	Data() string
  RBlocks() []*EventCell
}

type Update struct {
	cat  string
	data string
  rblocks []*EventCell
}

func (u Update) Category() string {
	return u.cat
}

func (u Update) Data() string {
	return u.data
}

func (u Update) RBlocks() []*EventCell {
  return u.rblocks
}

func NewUpdate(cat string, data string) UpdateEvent {
	return Update{cat: cat, data: data}
}

func NewRBUpdate(cat string, rblocks []*EventCell) UpdateEvent {
  return Update{cat: cat, rblocks: rblocks}
}
