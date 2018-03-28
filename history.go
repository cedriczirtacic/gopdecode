package main

import (
	"bufio"
	"fmt"
	"os"
)

/* config values */
const MAX_HIST = 200
const FILENAME = ".gop_history"

/* history structure */
type History struct {
	file *os.File
	cmds []string
	pos  int
}

var whitelist []string = []string{
	"q",
	"quit",
	"history",
}

func NewHist() (*History, error) {
	var (
		path string
		err  error
	)

	h := &History{}
	path = os.Getenv("HOME")
	if path == "" {
		path = os.TempDir()
		if path == "" {
			path = "."
		}
	}
	filename := fmt.Sprintf("%s/%s", path, FILENAME)
	h.file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0700)
	return h, err
}

func (h *History) Up() *string {
	i := h.pos
	if h.pos >= 1 {
		h.pos--
	}
	return &h.cmds[i]
}
func (h *History) Down() *string {
	if h.pos < (len(h.cmds) - 1) {
		h.pos++
	}
	i := h.pos
	return &h.cmds[i]
}

func (h *History) ResetPos() int {
	h.pos = len(h.cmds) - 1
	return h.pos
}

func (h *History) GetCmd(n int) *string {
	return &h.cmds[n]
}

func (h *History) AppendCmd(c string) int {
	/* whitelist commands */
	for _, cmd := range whitelist {
		if cmd == c {
			return 0
		}
	}

	if len(h.cmds) == MAX_HIST {
		h.cmds = append(h.cmds[1:], c)
	} else {
		h.cmds = append(h.cmds, c)
	}
	return len(h.cmds)
}

func (h *History) Search(c string) string {
	return c
}

func (h *History) Populate() error {
	scan := bufio.NewScanner(h.file)
	for i := 0; scan.Scan(); i++ {
		if i == MAX_HIST {
			break
		}
		h.cmds = append(h.cmds, scan.Text())
	}

	/* set the position to the last cmd */
	h.pos = len(h.cmds)
	if h.pos > 0 {
		h.pos--
	}

	return scan.Err()
}

func (h *History) Save() {
	defer h.file.Close()

	h.file.Truncate(0)
	for _, cmd := range h.cmds {
		h.file.WriteString(fmt.Sprintf("%s\n", cmd))
	}
}

func (h *History) PrintHistory() {
	if len(h.cmds) > 0 {
		for i, c := range h.cmds {
			fmt.Printf(" %d\t%s\n", i, c)
		}
	} else {
		println("empty\n")
	}
}
