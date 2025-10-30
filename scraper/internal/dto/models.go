package dto

type ListResponse struct {
	Result []struct {
		ID                  string `json:"id"`
		DevelopedDepartment struct {
			ImageID     string `json:"imageId"`
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"developedDepartment"`
		DeveloperUser                     interface{} `json:"developerUser"`
		ResponsibleEmployee               interface{} `json:"responsibleEmployee"`
		ProjectID                         string      `json:"projectId"`
		CreationDate                      string      `json:"creationDate"`
		Title                             string      `json:"title"`
		Stage                             string      `json:"stage"`
		Status                            string      `json:"status"`
		RegulatoryImpact                  string      `json:"regulatoryImpact"`
		StartPublicDiscussion             string      `json:"startPublicDiscussion"`
		EndPublicDiscussion               string      `json:"endPublicDiscussion"`
		StartParallelPublicDiscussion     interface{} `json:"startParallelPublicDiscussion"`
		EndParallelPublicDiscussion       interface{} `json:"endParallelPublicDiscussion"`
		Deadline                          interface{} `json:"deadline"`
		PublicationDate                   interface{} `json:"publicationDate"`
		StartPublicNotificationDiscussion interface{} `json:"startPublicNotificationDiscussion"`
		EndPublicNotificationDiscussion   interface{} `json:"endPublicNotificationDiscussion"`
		StartPublicTextDiscussion         interface{} `json:"startPublicTextDiscussion"`
		EndPublicTextDiscussion           interface{} `json:"endPublicTextDiscussion"`
		ImportantForRegions               bool        `json:"importantForRegions"`
		SupervisoryActivities             bool        `json:"supervisoryActivities"`
		StartPublicGradeConsultations     interface{} `json:"startPublicGradeConsultations"`
		EndPublicGradeConsultations       interface{} `json:"endPublicGradeConsultations"`
		Guillotine                        bool        `json:"guillotine"`
		Procedure                         struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"procedure"`
		ProjectType interface{} `json:"projectType"`
		KeyWords    interface{} `json:"keyWords"`
		Okveds      []struct {
			ImageID     string `json:"imageId"`
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"okveds"`
		Workflow             interface{} `json:"workflow"`
		NotificationList     interface{} `json:"notificationList"`
		ContactList          interface{} `json:"contactList"`
		Developers           interface{} `json:"developers"`
		ReasonForDevelopment interface{} `json:"reasonForDevelopment"`
		LinkedNpa            interface{} `json:"linkedNpa"`
		Hidden               bool        `json:"hidden"`
		IsOrv                interface{} `json:"isOrv"`
		NpaStatistics        struct {
			Views       int    `json:"views"`
			Rating      int    `json:"rating"`
			Comments    int    `json:"comments"`
			Description string `json:"description"`
		} `json:"npaStatistics"`
		NpaDiscussionStat interface{} `json:"npaDiscussionStat"`
		NpaRatings        interface{} `json:"npaRatings"`
	} `json:"result"`
	Count      int `json:"count"`
	TotalCount int `json:"totalCount"`
	Page       int `json:"page"`
}
