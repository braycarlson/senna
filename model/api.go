package model

type (
	/*
		Method: GET
	*/
	Champions map[string]struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	}

	/*
		Method: GET
	*/
	Runes struct {
		Primary   float64
		Secondary float64
		Runes     []float64
	}

	/*
		Method: GET
	*/
	Itemset struct {
		AccountId string `json:"accountId"`
		ItemSets  []struct {
			UID      string `json:"uid"`
			Title    string `json:"title"`
			Sortrank int    `json:"sortrank"`
			Type     string `json:"type"`
			Map      string `json:"map"`
			Mode     string `json:"mode"`
			Blocks   []struct {
				Items []struct {
					ID    string `json:"id"`
					Count int    `json:"count"`
				} `json:"items"`
				Type string `json:"type"`
			} `json:"blocks"`
			AssociatedMaps      []int         `json:"associatedMaps"`
			AssociatedChampions []string      `json:"associatedChampions"`
			PreferredItemSlots  []interface{} `json:"preferredItemSlots"`
		} `json:"itemSets"`
	}

	Itemsets map[string]Itemset

	/*
		Method: GET
	*/
	Runepage struct {
		AutoModifiedSelections [1]int    `json:"autoModifiedSelections"`
		Current                bool      `json:"current"`
		ID                     float64   `json:"id"`
		IsActive               bool      `json:"isActive"`
		IsDeletable            bool      `json:"isDeletable"`
		IsEditable             bool      `json:"isEditable"`
		IsValid                bool      `json:"isValid"`
		LastModified           float64   `json:"lastModified"`
		Name                   string    `json:"name"`
		Order                  float64   `json:"order"`
		PrimaryStyleID         float64   `json:"primaryStyleId"`
		SelectedPerkIDs        []float64 `json:"selectedPerkIds"`
		SubStyleID             float64   `json:"subStyleId"`
	}

	Runepages map[string]Runepage

	/*
		Method: GET
	*/
	SkillOrders map[string][]string

	SkillOrder map[string][]string
)
