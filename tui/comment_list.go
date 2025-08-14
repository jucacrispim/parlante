// Copyright 2025 Juca Crispim <juca@poraodojuca.dev>

// This file is part of parlante.

// parlante is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// parlante is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with parlante. If not, see <http://www.gnu.org/licenses/>.

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

// GetAddScreen returns the list screen because there is no add comment
// screen and it is easier just return the list screen
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
