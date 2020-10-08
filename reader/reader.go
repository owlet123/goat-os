package reader

import (
	"context"
	"time"

	"github.com/goat-project/goat-os/resource"

	networkReader "github.com/goat-project/goat-os/resource/network/reader"
	serverReader "github.com/goat-project/goat-os/resource/server/reader"
	storageReader "github.com/goat-project/goat-os/resource/storage/reader"

	"golang.org/x/time/rate"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/rafaeljesus/retry-go"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"

	"github.com/goat-project/goat-os/constants"
)

// Reader structure to list resources and retrieve info for specific resource from Openstack.
type Reader struct {
	client      *gophercloud.ServiceClient
	rateLimiter *rate.Limiter
	timeout     time.Duration
}

type resourcesReaderI interface {
	ReadResources(*gophercloud.ServiceClient) pagination.Pager
}

const attempts = 3
const sleepTime = time.Second * 1

// CreateReader creates reader with gophercloud client, rate limiter and timeout.
func CreateReader(client *gophercloud.ServiceClient, limiter *rate.Limiter) *Reader {
	if client == nil {
		log.WithFields(log.Fields{"error": "client is empty"}).Fatal("error create ServerByID")
	}

	log.WithFields(log.Fields{"attempts": attempts, "sleepTime": sleepTime}).Debug("ServerByID created " +
		"with given settings for number of iterations for unsuccessful calls and " +
		"sleep time between the calls")

	return &Reader{
		client:      client,
		rateLimiter: limiter,
		timeout:     viper.GetDuration(constants.CfgOpenstackTimeout),
	}
}

func (r *Reader) readResources(rri resourcesReaderI) (pagination.Pager, error) {
	var pager pagination.Pager
	var err error

	err = retry.Do(func() error {
		if err = r.rateLimiter.Wait(context.Background()); err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("error list resources")
		}

		pager = rri.ReadResources(r.client)

		return err
	}, attempts, sleepTime)

	return pager, err
}

// ListAllServers lists all servers from Openstack.
func (r *Reader) ListAllServers() (pagination.Pager, error) {
	return r.readResources(&serverReader.Servers{})
}

// ListAllUsers lists all users from Openstack.
func (r *Reader) ListAllUsers() (pagination.Pager, error) {
	return r.readResources(&resource.UserReader{})
}

// ListAllImages lists all images from Openstack.
func (r *Reader) ListAllImages() (pagination.Pager, error) {
	return r.readResources(&storageReader.Image{})
}

// ListAllShares lists all shares from Openstack.
func (r *Reader) ListAllShares() (pagination.Pager, error) {
	return r.readResources(&storageReader.Share{})
}

// FloatingIPs lists all floating ips.
func (r *Reader) FloatingIPs(id string) (pagination.Pager, error) {
	return r.readResources(&networkReader.FloatingIP{
		TenantID: id,
	})
}