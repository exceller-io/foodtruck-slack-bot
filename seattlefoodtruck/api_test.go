package seattlefoodtruck

import "testing"

func TestGetLocationEvents(t *testing.T) {
	p, _ := NewProxy("https://www.seattlefoodtruck.com")
	req := NewLocationEventsRequest(44, 1)

	r, err := p.GetLocationEvents(&req)
	if err != nil {
		t.Errorf("Expected '%v' got '%v'", r, nil)
	}
	t.Logf("Got %v \n", r)
}

func TestGetNeighborHoods(t *testing.T) {
	p, _ := NewProxy("https://www.seattlefoodtruck.com")

	n, err := p.GetNeighborhoods()
	if err != nil {
		t.Errorf("Error occurred ")
	}
	t.Logf("Got %v \n", n)
}

func TestGetLocations(t *testing.T) {
	p, _ := NewProxy("https://www.seattlefoodtruck.com")
	req := LocationRequest{
		Page:         1,
		Neighborhood: "bellevue",
	}
	lr, err := p.GetLocations(&req)
	if err != nil {
		t.Errorf("Expected %v got %v", lr, nil)
	}
	t.Logf("Got %v \n", lr)
}
