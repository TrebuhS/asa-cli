package services

import (
	"fmt"

	"github.com/trebuhs/asa-cli/internal/api"
	"github.com/trebuhs/asa-cli/internal/models"
)

type CampaignService struct {
	Client *api.Client
}

func NewCampaignService(client *api.Client) *CampaignService {
	return &CampaignService{Client: client}
}

func (s *CampaignService) List(limit, offset int) ([]models.Campaign, *models.PageDetail, error) {
	path := fmt.Sprintf("/campaigns?limit=%d&offset=%d", limit, offset)
	var campaigns []models.Campaign
	page, err := s.Client.Get(path, &campaigns)
	return campaigns, page, err
}

func (s *CampaignService) Get(id int64) (*models.Campaign, error) {
	var campaign models.Campaign
	_, err := s.Client.Get(fmt.Sprintf("/campaigns/%d", id), &campaign)
	return &campaign, err
}

func (s *CampaignService) Find(selector models.Selector) ([]models.Campaign, *models.PageDetail, error) {
	var campaigns []models.Campaign
	page, err := s.Client.Post("/campaigns/find", &selector, &campaigns)
	return campaigns, page, err
}

func (s *CampaignService) FindAll(selector models.Selector) ([]models.Campaign, error) {
	return api.PaginatedFetcher[models.Campaign](s.Client, "/campaigns/find", selector)
}

func (s *CampaignService) Create(campaign *models.Campaign) (*models.Campaign, error) {
	var created models.Campaign
	_, err := s.Client.Post("/campaigns", campaign, &created)
	return &created, err
}

func (s *CampaignService) Update(id int64, update *models.CampaignUpdate) (*models.Campaign, error) {
	var updated models.Campaign
	req := &models.UpdateCampaignRequest{Campaign: update}
	_, err := s.Client.Put(fmt.Sprintf("/campaigns/%d", id), req, &updated)
	return &updated, err
}

func (s *CampaignService) Delete(id int64) error {
	return s.Client.Delete(fmt.Sprintf("/campaigns/%d", id))
}
