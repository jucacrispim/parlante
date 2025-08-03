package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

func TestCommentItem(t *testing.T) {
	c := parlante.NewClientStorageInMemory()
	client, _, _ := c.CreateClient("test client")
	ds := parlante.NewClientDomainStorageInMemory()
	domain, _ := ds.AddClientDomain(client, "bla.net")
	cos := parlante.NewCommentStorageInMemory()
	comment, _ := cos.CreateComment(client, domain, "zé", "test comment", "http://bla.net")

	defer func() {
		cos.RemoveComment(comment)
	}()

	item := CommentItem{Comment: comment}

	if item.Title() != comment.Author {
		t.Fatalf("Bad title for item %s", item.Title())
	}

	if item.Description() != fmt.Sprintf("url: %s", comment.PageURL) {
		t.Fatalf("Bad description for item %s", item.Description())
	}

	if item.FilterValue() != comment.PageURL {
		t.Fatalf("Bad filter value for item %s", item.FilterValue())
	}
}

func TestCommentListScreen(t *testing.T) {

	c := parlante.NewClientStorageInMemory()
	cd := parlante.NewClientDomainStorageInMemory()
	comm := parlante.NewCommentStorageInMemory()
	main := newMainScreen(&c, &cd, &comm)

	c1, _, _ := c.CreateClient("a client")
	d1, _ := cd.AddClientDomain(c1, "domain.net")
	comm1, _ := comm.CreateComment(c1, d1, "zé", "the comment", "http://bla.net")
	comm2, _ := comm.CreateComment(c1, d1, "zé", "the other comment", "http://bla.net")
	var tests = []struct {
		testName string
		screenFn func() AddRemoveItemScreen
		msgFn    func(AddRemoveItemScreen) tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test load comments",
			func() AddRemoveItemScreen {
				return newCommentListScreen(&main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return m.Init()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model loading comments")
				}
				view := nm.View()
				if !strings.Contains(view, comm1.Author) ||
					!strings.Contains(view, comm2.Author) {
					t.Fatalf("comments not loaded %s", view)
				}

			},
		},
		{
			"test load comments with error",
			func() AddRemoveItemScreen {
				comm.ForceListError(true)
				return newCommentListScreen(&main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return m.Init()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				comm.ForceListError(false)
				nm, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model loading comments")
				}
				if nm.err == nil {
					t.Fatalf("No error with load comments error")
				}
			},
		},
		{
			"test GetAddScreen",
			func() AddRemoveItemScreen {
				return newCommentListScreen(&main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model for add client")
				}

			},
		},
		{
			"test GetRemoveScreen",
			func() AddRemoveItemScreen {
				s := newCommentListScreen(&main)
				items := s.Init()()
				i := items.(ItemListMsg)
				s.List.SetItems(i.Items)
				s.List.CursorDown()
				return s
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(removeCommentScreen)
				if !ok {
					t.Fatalf("bad model for remove comment")
				}

				if nm.Comment.PageURL != comm2.PageURL {
					t.Fatalf("bad comment on remove")
				}

			},
		},
		{
			"test GetPreviousScreen",
			func() AddRemoveItemScreen {
				return newCommentListScreen(&main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(mainScreen)
				if !ok {
					t.Fatalf("bad model for previous screen")
				}

			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			screen := test.screenFn()
			msg := test.msgFn(screen)
			m, cmd := screen.Update(msg)
			test.checkFn(m, cmd)
		})
	}
}
