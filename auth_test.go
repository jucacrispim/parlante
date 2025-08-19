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

package parlante

import (
	"errors"
	"testing"
)

type authTestStorage struct {
	ClientStorageInMemory
	client Client
	err    error
}

func (s authTestStorage) GetClientByUUID(uuid string) (Client, error) {
	return s.client, s.err
}

func newAuthTestStorage(c Client, err error) authTestStorage {
	cs := NewClientStorageInMemory()
	s := authTestStorage{
		ClientStorageInMemory: cs,
		client:                c,
		err:                   err,
	}
	return s
}

func TestAuthClient(t *testing.T) {

	type setupFn func() (ClientStorage, string, string)

	var tests = []struct {
		testName string
		setup    setupFn
		hasErr   bool
		err      error
	}{
		{
			"auth with error getting by uuid",
			func() (ClientStorage, string, string) {
				c, key, _ := NewClient("a client")
				s := newAuthTestStorage(c, errors.New("bad get by uuid"))
				return s, c.UUID, key
			},
			true,
			NO_CLIENT_ERR,
		},
		{
			"auth with bad key",
			func() (ClientStorage, string, string) {
				c, _, _ := NewClient("a client")
				s := newAuthTestStorage(c, nil)
				return s, c.UUID, "bad key"
			},
			true,
			INVALID_CREDS_ERR,
		},

		{
			"auth ok",
			func() (ClientStorage, string, string) {
				c, key, _ := NewClient("a client")
				s := newAuthTestStorage(c, nil)
				return s, c.UUID, key
			},
			false,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			s, uuid, key := test.setup()
			_, err := AuthClient(s, uuid, key)
			if test.hasErr && !errors.Is(err, test.err) {
				t.Fatalf("Bad err %+v", err)
			}
			if !test.hasErr && err != nil {
				t.Fatalf("error %s", err.Error())
			}
		})
	}
}
