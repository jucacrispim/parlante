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

import "errors"

var INVALID_CREDS_ERR error = errors.New("invald creds")
var NO_CLIENT_ERR error = errors.New("no client")

func AuthClient(s ClientStorage, uuid string, key string) (Client, error) {
	c, err := s.GetClientByUUID(uuid)
	if err != nil {
		return Client{}, NO_CLIENT_ERR
	}
	encr, err := HashStr(key)
	if err != nil {
		return Client{}, err
	}
	if encr != c.Key {
		return Client{}, INVALID_CREDS_ERR
	}
	return c, nil
}
