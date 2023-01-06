package generator

import (
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const header = `//go:generate go run ../generate/main.go -target $GOFILE
// THIS FILE IS GENERATED DO NOT EDIT
package minikube

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var (
	clusterSchema = map[string]*schema.Schema{
		"cluster_name": {
			Type:					schema.TypeString,
			Optional:			true,
			ForceNew:			true,
			Description:	"The name of the minikube cluster",
			Default:			"terraform-provider-minikube",
		},

		"nodes": {
			Type:					schema.TypeInt,
			Optional:			true,
			ForceNew:			true,
			Description:	"Amount of nodes in the cluster",
			Default:			1,
		},

		"client_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "client key for cluster",
			Sensitive:   true,
		},

		"client_certificate": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "client certificate used in cluster",
			Sensitive:   true,
		},

		"cluster_ca_certificate": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "certificate authority for cluster",
			Sensitive:   true,
		},

		"host": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "the host name for the cluster",
		},
`

func TestStringProperty(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("Version 999", nil)
	mockMinikube.EXPECT().GetStartHelpText(gomock.Any()).Return(`
--test='test-value':
	I am a great test description

--test2='test-value2':
	I am a great test2 description
	`, nil)
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	schema, err := builder.Build()
	assert.NoError(t, err)
	assert.Equal(t, header+`
		"test": {
			Type:					schema.TypeString,
			Description:	"I am a great test description",
			
			Optional:			true,
			ForceNew:			true,
			
			Default:	"test-value",
		},
	
		"test2": {
			Type:					schema.TypeString,
			Description:	"I am a great test2 description",
			
			Optional:			true,
			ForceNew:			true,
			
			Default:	"test-value2",
		},
	
	}
)

func GetClusterSchema() map[string]*schema.Schema {
	return clusterSchema
}
	`, schema)
}

func TestIntProperty(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("Version 999", nil)
	mockMinikube.EXPECT().GetStartHelpText(gomock.Any()).Return(`
--test=123:
	I am a great test description

	`, nil)
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	schema, err := builder.Build()
	assert.NoError(t, err)
	assert.Equal(t, header+`
		"test": {
			Type:					schema.TypeInt,
			Description:	"I am a great test description",
			
			Optional:			true,
			ForceNew:			true,
			
			Default:	123,
		},
	
	}
)

func GetClusterSchema() map[string]*schema.Schema {
	return clusterSchema
}
	`, schema)
}

func TestTimeProperty(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("Version 999", nil)
	mockMinikube.EXPECT().GetStartHelpText(gomock.Any()).Return(`
--test=6m0s:
	I am a great test description

	`, nil)
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	schema, err := builder.Build()
	assert.NoError(t, err)
	assert.Equal(t, header+`
		"test": {
			Type:					schema.TypeInt,
			Description:	"I am a great test description (Configured in minutes)",
			
			Optional:			true,
			ForceNew:			true,
			
			Default:	6,
		},
	
	}
)

func GetClusterSchema() map[string]*schema.Schema {
	return clusterSchema
}
	`, schema)
}

func TestBoolProperty(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("Version 999", nil)
	mockMinikube.EXPECT().GetStartHelpText(gomock.Any()).Return(`
--test=true:
	I am a great test description

	`, nil)
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	schema, err := builder.Build()
	assert.NoError(t, err)
	assert.Equal(t, header+`
		"test": {
			Type:					schema.TypeBool,
			Description:	"I am a great test description",
			
			Optional:			true,
			ForceNew:			true,
			
			Default:	true,
		},
	
	}
)

func GetClusterSchema() map[string]*schema.Schema {
	return clusterSchema
}
	`, schema)
}

func TestArrayProperty(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("Version 999", nil)
	mockMinikube.EXPECT().GetStartHelpText(gomock.Any()).Return(`
--test=[]:
	I am a great test description

	`, nil)
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	schema, err := builder.Build()
	assert.NoError(t, err)
	assert.Equal(t, header+`
		"test": {
			Type:					schema.TypeList,
			Description:	"I am a great test description",
			
			Optional:			true,
			ForceNew:			true,
			
			Elem: &schema.Schema{
				Type:	schema.TypeString,
			},
			
		},
	
	}
)

func GetClusterSchema() map[string]*schema.Schema {
	return clusterSchema
}
	`, schema)
}

func TestOutputProperty(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("Version 999", nil)
	mockMinikube.EXPECT().GetStartHelpText(gomock.Any()).Return(`
--host=123:
	I am a great test description

	`, nil)
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	schema, err := builder.Build()
	assert.NoError(t, err)
	assert.Equal(t, header+`
		"host": {
			Type:					schema.TypeInt,
			Description:	"I am a great test description",
			
			Computed:			true,
			
			Default:	123,
		},
	
	}
)

func GetClusterSchema() map[string]*schema.Schema {
	return clusterSchema
}
	`, schema)
}

func TestOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("Version 999", nil)
	mockMinikube.EXPECT().GetStartHelpText(gomock.Any()).Return(`
--memory='':
	I am a great test description

	`, nil)
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	schema, err := builder.Build()
	assert.NoError(t, err)
	assert.Equal(t, header+`
		"memory": {
			Type:					schema.TypeString,
			Description:	"Amount of RAM to allocate to Kubernetes (format: <number>[<unit>], where unit = b, k, m or g)",
			
			Optional:			true,
			ForceNew:			true,
			
			Default:	"4000mb",
		},
	
	}
)

func GetClusterSchema() map[string]*schema.Schema {
	return clusterSchema
}
	`, schema)
}

func TestPropertyFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("Version 999", nil)
	mockMinikube.EXPECT().GetStartHelpText(gomock.Any()).Return(`
--test=asdfasdf:
	I am a great test description

	`, nil)
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	_, err := builder.Build()
	assert.Error(t, err)
}

func TestNullDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("Version 999", nil)
	mockMinikube.EXPECT().GetStartHelpText(gomock.Any()).Return(`
--test=:
	I am a great test description

	`, nil)
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	_, err := builder.Build()
	assert.NoError(t, err)
}

func TestMinikubeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMinikube := NewMockMinikubeBinary(ctrl)
	mockMinikube.EXPECT().GetVersion(gomock.Any()).Return("", errors.New("could not find minikube binary"))
	builder := NewSchemaBuilder("fake.go", mockMinikube)
	_, err := builder.Build()
	assert.Error(t, err)
}