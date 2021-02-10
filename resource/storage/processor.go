package storage

import (
	"sync"

	"github.com/goat-project/goat-os/auth"
	"github.com/goat-project/goat-os/constants"
	"github.com/goat-project/goat-os/reader"
	"github.com/goat-project/goat-os/resource"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"

	log "github.com/sirupsen/logrus"
)

// Processor to process storage data.
type Processor struct {
	reader reader.Reader
}

// CreateProcessor creates Processor to manage reading from Openstack.
func CreateProcessor(r *reader.Reader) *Processor {
	if r == nil {
		log.WithFields(log.Fields{}).Error(constants.ErrCreateProcReaderNil)
		return nil
	}

	return &Processor{
		reader: *r,
	}
}

func (p *Processor) createComputeReader(osClient *gophercloud.ProviderClient) {
	cClient, err := auth.CreateComputeV2ServiceClient(osClient)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("unable to create Compute V2 service client")
		return
	}

	p.reader = *reader.CreateReader(cClient)
}

func (p *Processor) createSharesReader(osClient *gophercloud.ProviderClient) {
	cClient, err := auth.CreateSharedFileSystemV2ServiceClient(osClient)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("unable to create Shared File System V2 service client")
		return
	}

	p.reader = *reader.CreateReader(cClient)
}

// Reader gets reader.
func (p *Processor) Reader() *reader.Reader {
	return &p.reader
}

// Process provides listing of the images with pagination.
func (p *Processor) Process(_ projects.Project, osClient *gophercloud.ProviderClient, read chan resource.Resource,
	wg *sync.WaitGroup) {
	defer wg.Done()

	p.processImages(osClient, read)
	p.processShares(osClient, read)
}

func (p *Processor) processImages(osClient *gophercloud.ProviderClient, read chan resource.Resource) {
	p.createComputeReader(osClient)

	imgs, err := p.reader.ListAllImages()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error list images")
		return
	}

	pages, err := imgs.AllPages() // todo add openstack pagination and wg
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error get image pages")
		return
	}

	s, err := images.ExtractImages(pages)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error extract images")
		return
	}

	for i := range s {
		read <- &s[i]
	}
}

func (p *Processor) processShares(osClient *gophercloud.ProviderClient, read chan resource.Resource) {
	p.createSharesReader(osClient)

	shrs, err := p.reader.ListAllShares()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error list shares")
		return
	}

	pages, err := shrs.AllPages() // todo add openstack pagination and wg
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error get share pages")
		return
	}

	s, err := shares.ExtractShares(pages)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error extract shares")
		return
	}

	for i := range s {
		read <- &s[i]
	}
}

// RetrieveInfo - only for ? relevant.
func (p *Processor) RetrieveInfo(fullInfo chan resource.Resource, wg *sync.WaitGroup, image resource.Resource) {
	defer wg.Done()

	if image == nil {
		log.WithFields(log.Fields{}).Debug("retrieve info no image")
		return
	}

	fullInfo <- image
}