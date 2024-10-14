// Generated by PMS #364
package dws

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
)

func DataSourceDwsLogicalClusterVolumes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDwsLogicalClusterVolumesRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the region in which to query the resource. If omitted, the provider-level region will be used.`,
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specified the ID of the cluster to which the logical cluster belongs.`,
			},
			"volumes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: `The list of the disk volumes corresponding to the logical cluster.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logical_cluster_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The name of the logical cluster.`,
						},
						"percentage": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The percentage of disk space used.`,
						},
						"usage": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The used capacity of the disk.`,
						},
						"total": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The total capacity of the disk.`,
						},
					},
				},
			},
		},
	}
}

type LogicalClusterVolumesDSWrapper struct {
	*schemas.ResourceDataWrapper
	Config *config.Config
}

func newLogicalClusterVolumesDSWrapper(d *schema.ResourceData, meta interface{}) *LogicalClusterVolumesDSWrapper {
	return &LogicalClusterVolumesDSWrapper{
		ResourceDataWrapper: schemas.NewSchemaWrapper(d),
		Config:              meta.(*config.Config),
	}
}

func dataSourceDwsLogicalClusterVolumesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	wrapper := newLogicalClusterVolumesDSWrapper(d, meta)
	lisLogCluVolRst, err := wrapper.ListLogicalClusterVolumes()
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	err = wrapper.listLogicalClusterVolumesToSchema(lisLogCluVolRst)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// @API DWS GET /v2/{project_id}/clusters/{cluster_id}/logical-clusters/volumes
func (w *LogicalClusterVolumesDSWrapper) ListLogicalClusterVolumes() (*gjson.Result, error) {
	client, err := w.NewClient(w.Config, "dws")
	if err != nil {
		return nil, err
	}

	uri := "/v2/{project_id}/clusters/{cluster_id}/logical-clusters/volumes"
	uri = strings.ReplaceAll(uri, "{cluster_id}", w.Get("cluster_id").(string))
	return httphelper.New(client).
		Method("GET").
		URI(uri).
		OffsetPager("volumes", "offset", "limit", 0).
		Request().
		Result()
}

func (w *LogicalClusterVolumesDSWrapper) listLogicalClusterVolumesToSchema(body *gjson.Result) error {
	d := w.ResourceData
	mErr := multierror.Append(nil,
		d.Set("region", w.Config.GetRegion(w.ResourceData)),
		d.Set("volumes", schemas.SliceToList(body.Get("volumes"),
			func(volumes gjson.Result) any {
				return map[string]any{
					"logical_cluster_name": volumes.Get("logical_cluster_name").Value(),
					"percentage":           volumes.Get("percent").Value(),
					"usage":                volumes.Get("usage").Value(),
					"total":                volumes.Get("total").Value(),
				}
			},
		)),
	)
	return mErr.ErrorOrNil()
}