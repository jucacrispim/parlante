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

import "github.com/jucacrispim/parlante"

var loc = parlante.GetDefaultLocale()

var MESSAGE_CLIENTS = loc.Get("Clients")
var MESSAGE_CLIENTS_SCREEN_DESCR = loc.Get("add / remove clients")
var MESSAGE_DOMAINS = loc.Get("Domains")
var MESSAGE_DOMAINS_SCREEN_DESCR = loc.Get("add / remove domains")
var MESSAGE_COMMENTS = loc.Get("Comments")
var MESSAGE_COMMENTS_SCREEN_DESCR = loc.Get("manage comments")
var MESSAGE_CHOOSE_ONE = loc.Get("Choose one")
var MESSAGE_ADD_CLIENT = loc.Get("Add new client")
var MESSAGE_CLIENT_NAME = loc.Get("client name")
var MESSAGE_REMOVE_CLIENT = loc.Get("Remove client")
var MESSAGE_REMOVE_CLIENT_CONFIRM = loc.Get(
	"Really want to remove client {{.name}}?")
var MESSAGE_DOMAIN_DESCRIPTION = loc.Get("client: {{.clientName}}")
var MESSAGE_CHOOSE_CLIENT = loc.Get("Choose a client")
var MESSAGE_DOMAIN_NAME = loc.Get("domain name")
var MESSAGE_NEW_DOMAIN_FOR = loc.Get("New domain for {{.clientName}}")
var MESSAGE_REMOVE_DOMAIN = loc.Get("Remove domain")
var MESSAGE_REMOVE_DOMAIN_CONFIRM = loc.Get(
	"Really want to remove domain {{.domain}}?")
var MESSAGE_REMOVE_COMMENT = loc.Get("Remove comment")
var MESSAGE_REMOVE_COMMENT_CONFIRM = loc.Get(
	"Really want to remove comment from {{.name}} at {{.url}}?")

var MESSAGE_KEY_HELP_ADD = loc.Get("add")
var MESSAGE_KEY_HELP_REMOVE = loc.Get("remove")
var MESSAGE_KEY_HELP_PREV_SCREEN = loc.Get("prev screen")
var MESSAGE_KEY_HELP_UP = loc.Get("up")
var MESSAGE_KEY_HELP_DOWN = loc.Get("down")
var MESSAGE_KEY_HELP_PREV_PAGE = loc.Get("prev page")
var MESSAGE_KEY_HELP_NEXT_PAGE = loc.Get("next page")
var MESSAGE_KEY_HELP_LIST_START = loc.Get("go to start")
var MESSAGE_KEY_HELP_LIST_END = loc.Get("go to end")
var MESSAGE_KEY_HELP_LIST_FILTER = loc.Get("filter")
var MESSAGE_KEY_HELP_LIST_CLEAR_FILTER = loc.Get("clear filter")
var MESSAGE_KEY_HELP_CANCEL = loc.Get("cancel")
var MESSAGE_KEY_HELP_CONFIRM = loc.Get("confirm")
var MESSAGE_KEY_HELP_APPLY_FILTER = loc.Get("apply filter")
var MESSAGE_KEY_HELP_MORE = loc.Get("more")
var MESSAGE_KEY_HELP_CLOSE_HELP = loc.Get("close help")
var MESSAGE_KEY_HELP_QUIT = loc.Get("quit")
var MESSAGE_KEY_HELP_SELECT = loc.Get("select")
