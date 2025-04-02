package main

type AdmissionControl struct {
	use_ac bool
	storage map[string]interface{}
}

func NewAdmissionControl(use_ac bool) *AdmissionControl {
	return &AdmissionControl{
		use_ac: use_ac,
		storage:    make(map[string]interface{}),
	}
}

func (ac *AdmissionControl) Admit(key string) bool {
	if !ac.use_ac {
		return true
	}

	if _, ok := ac.storage[key]; !ok {
		ac.storage[key] = true
		return false
	}

	return true
}

func (ac *AdmissionControl) Size() int {
	return len(ac.storage)
}
