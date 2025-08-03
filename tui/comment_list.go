package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type CommentItem struct {
	Comment parlante.Comment
}

func (i CommentItem) Title() string { return i.Comment.Author }
func (i CommentItem) Description() string {
	return fmt.Sprintf("url: %s", i.Comment.PageURL)
}
func (i CommentItem) FilterValue() string { return i.Comment.PageURL }

type CommentListNavigation struct {
	MainScreen *mainScreen
}

func (n CommentListNavigation) GetAddScreen() tea.Model {
	s := newCommentListScreen(n.MainScreen)
	return s
}

func (n CommentListNavigation) GetRemoveScreen(item list.Item) tea.Model {
	i := item.(CommentItem)
	s := newRemoveCommentScreen(*n.MainScreen, i.Comment)
	return s
}

func (n CommentListNavigation) GetPreviousScreen() tea.Model {
	return *n.MainScreen
}

type CommentLoader struct {
	Storage parlante.CommentStorage
}

func (l CommentLoader) Load() tea.Cmd {
	return func() tea.Msg {
		Comments, err := l.Storage.ListComments(parlante.CommentsFilter{})

		if err != nil {
			msg := ItemListMsg{
				Err: err,
			}
			return msg
		}

		items := make([]list.Item, 0)
		for _, c := range Comments {
			item := CommentItem{
				Comment: c,
			}
			items = append(items, item)
		}
		msg := ItemListMsg{
			Items: items,
			Err:   nil,
		}
		return msg
	}
}

func newCommentListScreen(mainScreen *mainScreen) AddRemoveItemScreen {

	nav := CommentListNavigation{
		MainScreen: mainScreen,
	}
	l := CommentLoader{
		Storage: mainScreen.CommentStorage,
	}
	h := mainScreen.header
	opts := ListOpts{
		Title:           MESSAGE_COMMENTS,
		ShowDescription: true,
		ShowStatusBar:   true,
		ShowHelp:        true,
	}
	s := NewAddRemoveItemScreen(&h, opts, nav, l.Load)
	return s
}
