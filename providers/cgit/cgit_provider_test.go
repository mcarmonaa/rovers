package cgit

import (
	"errors"
	"io"
	"sync"

	"github.com/src-d/rovers/core"
	. "gopkg.in/check.v1"
)

type CgitProviderSuite struct {
}

var _ = Suite(&CgitProviderSuite{})

func (s *CgitProviderSuite) SetUpTest(c *C) {
	core.NewClient(cgitProviderName).DropDatabase()
}

func (s *CgitProviderSuite) newProvider(cgitUrls ...string) *provider {

	return &provider{
		cgitCollection: initializeCollection(),
		discoverer:     &dummyDiscoverer{cgitUrls},
		scrapers:       []*scraper{},
		mutex:          &sync.Mutex{},
		lastRepo:       nil,
	}
}

func (s *CgitProviderSuite) TestCgitProvider_WhenFinishScraping(c *C) {
	provider := s.newProvider("https://a3nm.net/git/")

	var err error = nil
	url := ""
	count := 0
	for err == nil {
		url, err = provider.Next()
		if err == nil {
			ackErr := provider.Ack(nil)
			c.Assert(ackErr, IsNil)
		}
		count++
	}

	c.Assert(count, Not(Equals), 0)
	c.Assert(url, Equals, "")
	c.Assert(err, Equals, io.EOF)

}

func (s *CgitProviderSuite) TestCgitProvider_WhenAckIsError(c *C) {
	provider := s.newProvider("https://a3nm.net/git/")

	urlOne, err := provider.Next()
	ackErr := provider.Ack(errors.New("OOPS"))
	c.Assert(err, IsNil)
	c.Assert(ackErr, IsNil)

	urlTwo, err := provider.Next()
	ackErr = provider.Ack(nil)
	c.Assert(err, IsNil)
	c.Assert(ackErr, IsNil)

	urlTree, err := provider.Next()
	c.Assert(err, IsNil)

	c.Assert(urlOne, Equals, urlTwo)
	c.Assert(urlTwo, Not(Equals), urlTree)
}

func (s *CgitProviderSuite) TestCgitProvider_NotSendAlreadySended(c *C) {
	provider := s.newProvider("https://a3nm.net/git/")

	urlOne, err := provider.Next()
	ackErr := provider.Ack(nil)
	c.Assert(err, IsNil)
	c.Assert(ackErr, IsNil)

	provider = s.newProvider("https://a3nm.net/git/")

	urlTwo, err := provider.Next()
	ackErr = provider.Ack(nil)
	c.Assert(err, IsNil)
	c.Assert(ackErr, IsNil)

	c.Assert(urlOne, Not(Equals), urlTwo)
}

func (s *CgitProviderSuite) TestCgitProvider_IterateAllUrls(c *C) {
	provider := s.newProvider("https://a3nm.net/git/", "https://ongardie.net/git/")
	maxIndex := 0
	for {
		_, err := provider.Next()
		if provider.currentScraperIndex > maxIndex {
			maxIndex = provider.currentScraperIndex
		}
		if err == io.EOF {
			break
		}
		c.Assert(err, IsNil)
		ackErr := provider.Ack(nil)
		c.Assert(ackErr, IsNil)
	}
	c.Assert(maxIndex, Equals, 1)
	c.Assert(provider.currentScraperIndex, Equals, 0)
	c.Assert(len(provider.scrapers), Equals, 0)
}

func (s *CgitProviderSuite) TestCgitProvider_ScrapersWithDifferentUrls(c *C) {
	provider := s.newProvider("https://a3nm.net/git/", "https://a3nm.net/git/", "https://ongardie.net/git/")
	_, err := provider.Next()
	c.Assert(err, IsNil)
	c.Assert(len(provider.scrapers), Equals, 2)
}

type dummyDiscoverer struct {
	urls []string
}

func (d *dummyDiscoverer) Discover() []string {
	return d.urls
}

func (d *dummyDiscoverer) Reset() {}