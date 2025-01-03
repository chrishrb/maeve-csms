package services

import (
	"fmt"
	"regexp"
)

type ExtractService interface {
	GetCountryCode(evseId string) (string, error)
	GetOperatorId(evseId string) (string, error)
	GetChargeStationId(evseId string) (string, error)
}

type EvseUIDService struct {
	pattern string
}

func NewEvseUIDService(pattern string) *EvseUIDService {
	return &EvseUIDService{
		pattern: pattern,
	}
}

func (s *EvseUIDService) GetChargeStationId(evseId string) (string, error) {
	regex := regexp.MustCompile(s.pattern)
	match := regex.FindStringSubmatch(evseId)

	if len(match) < 4 {
		return "", fmt.Errorf("invalid EVSE ID format, could not extract charge point id: %s", evseId)
	}

	chargePointId := match[3]
	return chargePointId, nil
}

func (s *EvseUIDService) GetOperatorId(evseId string) (string, error) {
	regex := regexp.MustCompile(s.pattern)
	match := regex.FindStringSubmatch(evseId)

	if len(match) < 4 {
		return "", fmt.Errorf("invalid EVSE ID format, could not extract operator id: %s", evseId)
	}

	chargePointId := match[2]
	return chargePointId, nil
}

func (s *EvseUIDService) GetCountryCode(evseId string) (string, error) {
	regex := regexp.MustCompile(s.pattern)
	match := regex.FindStringSubmatch(evseId)

	if len(match) < 4 {
		return "", fmt.Errorf("invalid EVSE ID format, could not extract country code: %s", evseId)
	}

	chargePointId := match[1]
	return chargePointId, nil
}
