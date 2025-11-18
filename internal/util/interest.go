package util

type SystemInterestSet struct {
	interests map[string]struct{}
}

func NewSystemInterestSet(systemInterests []string) *SystemInterestSet {
	set := make(map[string]struct{})
	for _, interest := range systemInterests {
		set[interest] = struct{}{}
	}
	return &SystemInterestSet{
		interests: set,
	}
}

func (s *SystemInterestSet) Contains(interest string) bool {
	_, ok := s.interests[interest]
	return ok
}
