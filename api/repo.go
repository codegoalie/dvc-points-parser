package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/codegoalie/dvc-points-parser/models"
	"github.com/gojek/heimdall/httpclient"
)

const timeout = 15 * time.Second

type Repo struct {
	baseURL string
}

type PointsReq struct {
	RoomTypeID int       `json:"room_type_id"`
	StayOn     time.Time `json:"stay_on"`
	Amount     int       `json:"amount"`
}

func New(baseURL string) Repo {
	return Repo{baseURL: baseURL}
}

func (r Repo) GetResorts() ([]models.Resort, error) {
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	resorts := []models.Resort{}
	res, err := client.Get(r.baseURL+"/resorts", nil)
	if err != nil {
		err = fmt.Errorf("failed to execute resorts request: %w", err)
		return resorts, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("failed to read resorts body: %w", err)
		return resorts, err
	}
	res.Body.Close()
	// body := []byte(resortsResponse)

	err = json.Unmarshal(body, &resorts)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal resorts response: %w \n\n%s", err, body)
		return resorts, err
	}

	return resorts, nil
}

func (r Repo) GetRoomTypes(resortID int) ([]models.Room, error) {
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	roomTypes := []models.Room{}
	res, err := client.Get(fmt.Sprintf("%s/resorts/%d/room-types", r.baseURL, resortID), nil)
	if err != nil {
		err = fmt.Errorf("failed to execute room type request: %w", err)
		return roomTypes, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("failed to read room types body: %w", err)
		return roomTypes, err
	}
	res.Body.Close()

	err = json.Unmarshal(body, &roomTypes)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal room types response: %w", err)
		return roomTypes, err
	}

	return roomTypes, nil
}

func (r Repo) CreatePoints(points []PointsReq) error {
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	raw, err := json.Marshal(&points)
	if err != nil {
		err = fmt.Errorf("failed to marshal create points payload: %w", err)
		return err
	}

	buf := bytes.NewBuffer(raw)
	_, err = client.Post(
		fmt.Sprintf("%s/points", r.baseURL),
		buf,
		http.Header(map[string][]string{
			"Content-Type": {"application/json"},
		}),
	)
	if err != nil {
		err = fmt.Errorf("failed to create points: %w", err)
		return err
	}

	return nil
}

const resortsResponse = `
[
	{
		"id": 1,
		"name": "Aulani, Disney Vacation Club Villas, Ko Olina, Hawaii",
		"abbreviation": "AUL",
		"code": "AULV"
	},
	{
		"id": 2,
		"name": "Disney's BoardWalk Villas",
		"abbreviation": "BWV",
		"code": "BWALK"
	},
	{
		"id": 3,
		"name": "Copper Creek Villas & Cabins at Disney's Wilderness Lodge",
		"abbreviation": "CCR",
		"code": "WCC"
	},
	{
		"id": 4,
		"name": "Disney's Animal Kingdom Villas - Jambo House",
		"abbreviation": "AKV",
		"code": "AKV"
	},
	{
		"id": 5,
		"name": "Disney's Animal Kingdom Villas - Kidani Village",
		"abbreviation": "AKV2",
		"code": "AKV2"
	},
	{
		"id": 6,
		"name": "Disney's Beach Club Villas",
		"abbreviation": "BCV",
		"code": "BCV"
	},
	{
		"id": 7,
		"name": "Bay Lake Tower at Disney's Contemporary Resort",
		"abbreviation": "BLT",
		"code": "BLT"
	},
	{
		"id": 8,
		"name": "Disney's Saratoga Springs Resort & Spa",
		"abbreviation": "SSR",
		"code": "SSR"
	},
	{
		"id": 9,
		"name": "The Villas at Disney's Grand Californian Hotel & Spa",
		"abbreviation": "VGC",
		"code": "GCAL"
	},
	{
		"id": 10,
		"name": "Disney's Hilton Head Island Resort",
		"abbreviation": "HHI",
		"code": "HILTN"
	},
	{
		"id": 11,
		"name": "Disney's Old Key West Resort",
		"abbreviation": "OKW",
		"code": "CLUB"
	},
	{
		"id": 12,
		"name": "Disney's Vero Beach Resort",
		"abbreviation": "VBR",
		"code": "VERO"
	},
	{
		"id": 13,
		"name": "Boulder Ridge Villas at Disney's Wilderness Lodge",
		"abbreviation": "VWL",
		"code": "VWL"
	},
	{
		"id": 14,
		"name": "Disney's Polynesian Villas & Bungalows",
		"abbreviation": "POLY",
		"code": "POLYV"
	},
	{
		"id": 15,
		"name": "Disney's Riviera Resort",
		"abbreviation": "RIV",
		"code": "RVA"
	},
	{
		"id": 16,
		"name": "The Villas at Disney's Grand Floridian Resort & Spa",
		"abbreviation": "VGF",
		"code": "VGF"
	}
]
`
