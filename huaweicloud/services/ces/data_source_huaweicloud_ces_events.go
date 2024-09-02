// Generated by PMS #321
package ces

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tidwall/gjson"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/helper/httphelper"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/helper/schemas"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func DataSourceCesEvents() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCesEventsRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the region in which to query the resource. If omitted, the provider-level region will be used.`,
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the event type.`,
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the event name.`,
			},
			"from": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the start time of the query.`,
			},
			"to": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  `Specifies the end time of the query.`,
				RequiredWith: []string{"from"},
			},
			"events": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: `The event records.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The event name.`,
						},
						"event_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The event type.`,
						},
						"event_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: `The number of occurrences of this event within the specified query time range.`,
						},
						"latest_occur_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The time when the event last occurred.`,
						},
						"latest_event_source": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The event source.`,
						},
					},
				},
			},
		},
	}
}

type EventsDSWrapper struct {
	*schemas.ResourceDataWrapper
	Config *config.Config
}

func newEventsDSWrapper(d *schema.ResourceData, meta interface{}) *EventsDSWrapper {
	return &EventsDSWrapper{
		ResourceDataWrapper: schemas.NewSchemaWrapper(d),
		Config:              meta.(*config.Config),
	}
}

func dataSourceCesEventsRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	wrapper := newEventsDSWrapper(d, meta)
	listEventsRst, err := wrapper.ListEvents()
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	err = wrapper.listEventsToSchema(listEventsRst)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// @API CES GET /V1.0/{project_id}/events
func (w *EventsDSWrapper) ListEvents() (*gjson.Result, error) {
	client, err := w.NewClient(w.Config, "ces")
	if err != nil {
		return nil, err
	}

	uri := "/V1.0/{project_id}/events"
	from := w.Get("from")
	to := w.Get("to")
	var startTime, endTime int64
	params := map[string]any{
		"event_type": w.Get("type"),
		"event_name": w.Get("name"),
	}

	if from != nil {
		startTime, err = parseTimeToTimestamp(from.(string))
		if err != nil {
			return nil, err
		}
		params["from"] = startTime * 1000
	}

	if to != nil {
		endTime, err = parseTimeToTimestamp(to.(string))
		if err != nil {
			return nil, err
		}
		params["to"] = endTime * 1000
	}

	params = utils.RemoveNil(params)
	return httphelper.New(client).
		Method("GET").
		URI(uri).
		Query(params).
		OffsetPager("events", "start", "limit", 100).
		Request().
		Result()
}

func (w *EventsDSWrapper) listEventsToSchema(body *gjson.Result) error {
	d := w.ResourceData
	mErr := multierror.Append(nil,
		d.Set("region", w.Config.GetRegion(w.ResourceData)),
		d.Set("events", schemas.SliceToList(body.Get("events"),
			func(events gjson.Result) any {
				return map[string]any{
					"event_name":          events.Get("event_name").Value(),
					"event_type":          events.Get("event_type").Value(),
					"event_count":         events.Get("event_count").Value(),
					"latest_occur_time":   w.setEveLatOccTime(events),
					"latest_event_source": events.Get("latest_event_source").Value(),
				}
			},
		)),
	)
	return mErr.ErrorOrNil()
}

func parseTimeToTimestamp(timeStr string) (int64, error) {
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func (*EventsDSWrapper) setEveLatOccTime(data gjson.Result) string {
	return utils.FormatTimeStampRFC3339(int64(data.Get("latest_occur_time").Float()/1000), true)
}