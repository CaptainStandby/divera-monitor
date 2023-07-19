package alarm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlarmHandler(t *testing.T) {

	tt := []struct {
		name       string
		method     string
		body       string
		statusCode int
		expect     *jsonAlarm
	}{
		{
			name:       "GET not allowed",
			method:     http.MethodGet,
			body:       "",
			statusCode: http.StatusMethodNotAllowed,
			expect:     nil,
		},
		{
			name:       "reject invalid data",
			method:     http.MethodPost,
			body:       "{ \"title\": 42 }",
			statusCode: http.StatusBadRequest,
			expect:     nil,
		},
		{
			name:   "Accept valid data 1",
			method: http.MethodPost,
			body: `{
				"id": 11253967,
				"foreign_id": "",
				"title": "TEST TEST TEST",
				"text": "",
				"address": "Bockholz 2, Winnemark, Germany",
				"lat": "54.60561010",
				"lng": "9.93120260",
				"priority": 0,
				"notification_type": 4,
				"cluster": [],
				"vehicle": [],
				"group": [],
				"user_cluster_relation": [530527],
				"ts_create": 1689757211,
				"ts_update": 1689757211
			}`,
			statusCode: http.StatusOK,
			expect: &jsonAlarm{
				ID:               11253967,
				ForeignID:        "",
				Title:            "TEST TEST TEST",
				Text:             "",
				Address:          "Bockholz 2, Winnemark, Germany",
				Lat:              "54.60561010",
				Lng:              "9.93120260",
				Priority:         0,
				NotificationType: 4,
				Created:          1689757211,
				Updated:          1689757211,
			},
		},
		{
			name:   "Accept valid data 2",
			method: http.MethodPost,
			body: `{
				"id": 11254160,
				"foreign_id": "",
				"title": "TEST TEST TEST",
				"text": "",
				"address": "",
				"lat": null,
				"lng": null,
				"priority": 1,
				"notification_type": 4,
				"cluster": [],
				"vehicle": [],
				"group": [],
				"user_cluster_relation": [530527],
				"ts_create": 1689758202,
				"ts_update": 1689758202
			}`,
			statusCode: http.StatusOK,
			expect: &jsonAlarm{
				ID:               11254160,
				ForeignID:        "",
				Title:            "TEST TEST TEST",
				Text:             "",
				Address:          "",
				Lat:              "",
				Lng:              "",
				Priority:         1,
				NotificationType: 4,
				Created:          1689758202,
				Updated:          1689758202,
			},
		},
		{
			name:   "Accept valid data 3",
			method: http.MethodPost,
			body: `{
				"id": 11253859,
				"foreign_id": "",
				"title": "TEST TEST TEST",
				"text": "",
				"address": "Bockholz 2, Winnemark, Germany",
				"lat": null,
				"lng": null,
				"priority": 0,
				"notification_type": 4,
				"cluster": [],
				"vehicle": [],
				"group": [],
				"user_cluster_relation": [530527],
				"ts_create": 1689756639,
				"ts_update": 1689756662
			}`,
			statusCode: http.StatusOK,
			expect: &jsonAlarm{
				ID:               11253859,
				ForeignID:        "",
				Title:            "TEST TEST TEST",
				Text:             "",
				Address:          "Bockholz 2, Winnemark, Germany",
				Lat:              "",
				Lng:              "",
				Priority:         0,
				NotificationType: 4,
				Created:          1689756639,
				Updated:          1689756662,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			var pushed *jsonAlarm
			handle(responseRecorder, request, func(ctx context.Context, ja *jsonAlarm) error {
				pushed = ja
				return nil
			})

			assert.Equal(t, tc.statusCode, responseRecorder.Code)
			assert.Equal(t, tc.expect, pushed)
		})
	}
}
