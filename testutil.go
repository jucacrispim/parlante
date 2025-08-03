package parlante

// notest

import "errors"

// A in memory database for tests
type ClientStorageInMemory struct {
	data          map[string]Client
	BadClientUUID string
	listError     bool
	removeError   bool
}

func (s ClientStorageInMemory) CreateClient(name string) (Client, string, error) {
	c, key, _ := NewClient(name)
	s.data[c.UUID] = c
	return c, key, nil
}

func (s ClientStorageInMemory) GetClientByUUID(uuid string) (Client, error) {
	if uuid == s.BadClientUUID {
		return Client{}, errors.New("Bad!")
	}
	c, ok := s.data[uuid]
	if !ok {
		return Client{}, nil
	}
	return c, nil
}

func (s ClientStorageInMemory) ListClients() ([]Client, error) {
	if s.listError {
		return nil, errors.New("error list client")
	}
	clients := make([]Client, 0)
	for _, v := range s.data {
		clients = append(clients, v)
	}
	return clients, nil
}

func (s ClientStorageInMemory) RemoveClient(uuid string) error {
	if s.removeError {
		return errors.New("error remove client")
	}
	delete(s.data, uuid)
	return nil
}

func (s *ClientStorageInMemory) ForceListError(f bool) {
	s.listError = f
}

func (s *ClientStorageInMemory) ForceRemoveError(f bool) {
	s.removeError = f
}

func NewClientStorageInMemory() ClientStorageInMemory {
	c := ClientStorageInMemory{}
	c.data = make(map[string]Client)
	u, _ := GenUUID4()
	c.BadClientUUID = u
	return c
}

type ClientDomainStorageInMemory struct {
	data        map[string]ClientDomain
	BadDomain   string
	listError   bool
	removeError bool
}

func (s ClientDomainStorageInMemory) AddClientDomain(c Client, domain string) (
	ClientDomain, error) {
	d := NewClientDomain(c, domain)
	key := c.UUID + "-" + d.Domain
	s.data[key] = d
	return d, nil
}

func (s ClientDomainStorageInMemory) RemoveClientDomain(c Client, domain string) error {
	if s.removeError {
		return errors.New("bad remove domain")
	}
	key := c.UUID + "-" + domain
	delete(s.data, key)
	return nil
}

func (s ClientDomainStorageInMemory) GetClientDomain(c Client, domain string) (
	ClientDomain, error) {
	if domain == s.BadDomain {
		return ClientDomain{}, errors.New("bad")
	}
	key := c.UUID + "-" + domain
	d, ok := s.data[key]
	if !ok {
		return ClientDomain{}, nil
	}
	return d, nil
}

func (s ClientDomainStorageInMemory) ListDomains() ([]ClientDomain, error) {
	if s.listError {
		return nil, errors.New("Bad list domain error!")
	}
	domains := make([]ClientDomain, 0)
	for _, v := range s.data {
		domains = append(domains, v)
	}
	return domains, nil

}
func (s *ClientDomainStorageInMemory) ForceListError(f bool) {
	s.listError = f
}

func (s *ClientDomainStorageInMemory) ForceRemoveError(f bool) {
	s.removeError = f
}

func NewClientDomainStorageInMemory() ClientDomainStorageInMemory {
	d := ClientDomainStorageInMemory{}
	d.data = make(map[string]ClientDomain)
	d.BadDomain = "bad.net"
	return d
}

type CommentStorageInMemory struct {
	data           map[string][]Comment
	clientComments map[int64][]Comment
	domainComments map[int64][]Comment
	pageComments   map[string][]Comment
	BadCommenter   string
	BadPage        string
	listError      bool
	removeError    bool
}

func (s CommentStorageInMemory) CreateComment(c Client, d ClientDomain,
	name string, content string, page_url string) (Comment, error) {
	if name == s.BadCommenter {
		return Comment{}, errors.New("bad")
	}

	comment := NewComment(c, d, name, content, page_url)
	s.data["all"] = append(s.data["all"], comment)
	s.clientComments[c.ID] = append(s.clientComments[c.ID], comment)
	s.domainComments[d.ID] = append(s.domainComments[d.ID], comment)
	s.pageComments[comment.PageURL] = append(
		s.pageComments[comment.PageURL], comment)

	return comment, nil
}

func (s CommentStorageInMemory) ListComments(filter CommentsFilter) (
	[]Comment, error) {

	if s.listError {
		return nil, errors.New("bad")
	}

	if filter.PageURL != nil && *filter.PageURL == s.BadPage {
		return []Comment{}, errors.New("bad")
	}

	if filter.ClientID != nil {
		return s.clientComments[*filter.ClientID], nil
	}
	if filter.DomainID != nil {
		return s.domainComments[*filter.DomainID], nil
	}

	if filter.PageURL != nil {
		return s.pageComments[*filter.PageURL], nil
	}
	return s.data["all"], nil
}

func (s CommentStorageInMemory) RemoveComment(comment Comment) error {
	if s.removeError {
		return errors.New("bad")
	}
	if len(s.data["all"]) < 1 {
		return nil
	}
	s.data["all"] = append(s.data["all"][:0], s.data["all"][1:]...)
	return nil
}

func (s CommentStorageInMemory) GetComment() Comment {
	if len(s.data["all"]) > 0 {
		return s.data["all"][0]
	}
	return Comment{}
}

func (s *CommentStorageInMemory) ForceListError(force bool) {
	s.listError = force
}

func (s *CommentStorageInMemory) ForceRemoveError(force bool) {
	s.removeError = force
}

func NewCommentStorageInMemory() CommentStorageInMemory {
	c := CommentStorageInMemory{}
	c.data = make(map[string][]Comment, 0)
	c.clientComments = make(map[int64][]Comment)
	c.domainComments = make(map[int64][]Comment)
	c.pageComments = make(map[string][]Comment)
	c.BadCommenter = "bad"
	c.BadPage = "http://bla.net/bad"
	return c
}
