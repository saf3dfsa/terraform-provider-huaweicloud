// Generated by PMS #425
package dms

import (
	"context"
	"strings"

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

func DataSourceDmsKafkaTopicProducers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDmsKafkaTopicProducersRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the region in which to query the resource. If omitted, the provider-level region will be used.`,
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specifies the instance ID.`,
			},
			"topic": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specifies the topic name.`,
			},
			"producers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: `Indicates the producer list.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"producer_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Indicates the producer address.`,
						},
						"broker_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Indicates the broker address.`,
						},
						"join_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Indicates the time when the broker was connected.`,
						},
					},
				},
			},
		},
	}
}

type KafkaTopicProducersDSWrapper struct {
	*schemas.ResourceDataWrapper
	Config *config.Config
}

func newKafkaTopicProducersDSWrapper(d *schema.ResourceData, meta interface{}) *KafkaTopicProducersDSWrapper {
	return &KafkaTopicProducersDSWrapper{
		ResourceDataWrapper: schemas.NewSchemaWrapper(d),
		Config:              meta.(*config.Config),
	}
}

func dataSourceDmsKafkaTopicProducersRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	wrapper := newKafkaTopicProducersDSWrapper(d, meta)
	lisTopProRst, err := wrapper.ListTopicProducers()
	if err != nil {
		return diag.FromErr(err)
	}

	err = wrapper.listTopicProducersToSchema(lisTopProRst)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	return nil
}

// @API Kafka GET /v2/{project_id}/kafka/instances/{instance_id}/topics/{topic}/producers
func (w *KafkaTopicProducersDSWrapper) ListTopicProducers() (*gjson.Result, error) {
	client, err := w.NewClient(w.Config, "dmsv2")
	if err != nil {
		return nil, err
	}

	uri := "/v2/{project_id}/kafka/instances/{instance_id}/topics/{topic}/producers"
	uri = strings.ReplaceAll(uri, "{instance_id}", w.Get("instance_id").(string))
	uri = strings.ReplaceAll(uri, "{topic}", w.Get("topic").(string))
	return httphelper.New(client).
		Method("GET").
		URI(uri).
		OffsetPager("producers", "offset", "limit", 0).
		Request().
		Result()
}

func (w *KafkaTopicProducersDSWrapper) listTopicProducersToSchema(body *gjson.Result) error {
	d := w.ResourceData
	mErr := multierror.Append(nil,
		d.Set("region", w.Config.GetRegion(w.ResourceData)),
		d.Set("producers", schemas.SliceToList(body.Get("producers"),
			func(producers gjson.Result) any {
				return map[string]any{
					"producer_address": producers.Get("producer_address").Value(),
					"broker_address":   producers.Get("broker_address").Value(),
					"join_time":        w.setProducersJoinTime(producers),
				}
			},
		)),
	)
	return mErr.ErrorOrNil()
}

func (*KafkaTopicProducersDSWrapper) setProducersJoinTime(data gjson.Result) string {
	return utils.FormatTimeStampRFC3339((data.Get("join_time").Int())/1000, true)
}