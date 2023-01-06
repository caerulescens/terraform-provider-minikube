//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock_minikube_binary.go -package=$GOPACKAGE
package generator

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type MinikubeBinary interface {
	GetVersion(ctx context.Context) (string, error)
	GetStartHelpText(ctx context.Context) (string, error)
}

type MinikubeHostBinary struct {
}

func (m *MinikubeHostBinary) GetVersion(ctx context.Context) (string, error) {
	return run(ctx, "version")
}

func (m *MinikubeHostBinary) GetStartHelpText(ctx context.Context) (string, error) {
	return run(ctx, "start", "--help")
}

var computedFields []string = []string{
	"addons",
	"apiserver_ips",
	"apiserver_names",
	"hyperkit_vsock_ports",
	"insecure_registry",
	"iso_url",
	"nfs_share",
	"ports",
	"registry_mirror",
	"client_key",
	"client_certificate",
	"cluster_ca_certificate",
	"host",
}

type SchemaOverride struct {
	Description string
	Default     string
	Type        SchemaType
}

var schemaOverrides map[string]SchemaOverride = map[string]SchemaOverride{
	"memory": {
		Default:     "4000mb",
		Description: "Amount of RAM to allocate to Kubernetes (format: <number>[<unit>], where unit = b, k, m or g)",
		Type:        String,
	},
	"cpus": {
		Default:     "2",
		Description: "Amount of CPUs to allocate to Kubernetes",
		Type:        Int,
	},
}

func run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "minikube", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

type SchemaEntry struct {
	Parameter   string
	Default     string
	Description string
	Type        SchemaType
	ArrayType   SchemaType
}

type SchemaBuilder struct {
	targetFile string
	minikube   MinikubeBinary
}

type SchemaType string

const (
	String SchemaType = "String"
	Int    SchemaType = "Int"
	Bool   SchemaType = "Bool"
	Array  SchemaType = "List"
)

func NewSchemaBuilder(targetFile string, minikube MinikubeBinary) *SchemaBuilder {
	return &SchemaBuilder{targetFile: targetFile, minikube: minikube}
}

func (s *SchemaBuilder) Build() (string, error) {
	minikubeVersion, err := s.minikube.GetVersion(context.Background())
	if err != nil {
		return "", errors.New("could not run minikube binary. please ensure that you have minikube installed and available")
	}

	log.Printf("building schema for minikube version: %s", minikubeVersion)

	help, err := s.minikube.GetStartHelpText(context.Background())
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(help))

	entries := make([]SchemaEntry, 0)

	currentEntry := SchemaEntry{}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "--") {
			currentEntry = loadParameter(line)
		} else if line != "" {
			currentEntry.Description += line
		} else if currentEntry.Parameter != "" {

			currentEntry.Description = strings.ReplaceAll(currentEntry.Description, "\\", "\\\\")
			currentEntry.Description = strings.ReplaceAll(currentEntry.Description, "\"", "\\\"")

			// Apply description override once we've built the description
			val, ok := schemaOverrides[currentEntry.Parameter]
			if ok {
				currentEntry.Description = val.Description
			}

			entries, err = addEntry(entries, currentEntry)
			if err != nil {
				return "", err
			}

			currentEntry.Parameter = ""
		}
	}

	schema := constructSchema(entries)

	return schema, err
}

func loadParameter(line string) SchemaEntry {
	schemaEntry := SchemaEntry{}
	schemaEntry.Description = ""
	seg := strings.Split(line, "=")
	schemaEntry.Parameter = strings.TrimPrefix(seg[0], "--")
	schemaEntry.Parameter = strings.Replace(schemaEntry.Parameter, "-", "_", -1)
	schemaEntry.Default = strings.TrimSuffix(seg[1], ":")
	schemaEntry.Type = getSchemaType(schemaEntry.Default)

	// Apply explicit overrides
	val, ok := schemaOverrides[schemaEntry.Parameter]
	if ok {
		schemaEntry.Default = val.Default
		schemaEntry.Type = val.Type
	}

	if schemaEntry.Type == String {
		schemaEntry.Default = strings.Trim(schemaEntry.Default, "'")
	}

	return schemaEntry
}

func addEntry(entries []SchemaEntry, currentEntry SchemaEntry) ([]SchemaEntry, error) {
	switch currentEntry.Type {
	case String:
		entries = append(entries, SchemaEntry{
			Parameter:   currentEntry.Parameter,
			Default:     fmt.Sprintf("\"%s\"", currentEntry.Default),
			Type:        currentEntry.Type,
			Description: currentEntry.Description,
		})
	case Bool:
		entries = append(entries, SchemaEntry{
			Parameter:   currentEntry.Parameter,
			Default:     currentEntry.Default,
			Type:        currentEntry.Type,
			Description: currentEntry.Description,
		})
	case Int:
		val, err := strconv.Atoi(currentEntry.Default)
		if err != nil {
			// is it a timestamp?
			time, err := time.ParseDuration(currentEntry.Default)
			if err != nil {
				return nil, err
			}
			val = int(time.Minutes())
			currentEntry.Description = fmt.Sprintf("%s (Configured in minutes)", currentEntry.Description)
		}
		entries = append(entries, SchemaEntry{
			Parameter:   currentEntry.Parameter,
			Default:     strconv.Itoa(val),
			Type:        currentEntry.Type,
			Description: currentEntry.Description,
		})
	case Array:
		entries = append(entries, SchemaEntry{
			Parameter:   currentEntry.Parameter,
			Type:        Array,
			ArrayType:   String,
			Description: currentEntry.Description,
		})
	}

	return entries, nil
}

func (s *SchemaBuilder) Write(schema string) error {
	return os.WriteFile(s.targetFile, []byte(schema), 0644)
}

func constructSchema(entries []SchemaEntry) string {

	header := `//go:generate go run ../generate/main.go -target $GOFILE
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

	body := ""
	for _, entry := range entries {
		extraParams := ""
		if !contains(computedFields, entry.Parameter) {
			extraParams = `
			Optional:			true,
			ForceNew:			true,
			`
		} else {
			extraParams = `
			Computed:			true,
			`
		}

		if entry.Type == Array {
			extraParams += fmt.Sprintf(`
			Elem: &schema.Schema{
				Type:	%s,
			},
			`, "schema.Type"+entry.ArrayType)
		} else {
			extraParams += fmt.Sprintf(`
			Default:	%s,`, entry.Default)
		}

		body = body + fmt.Sprintf(`
		"%s": {
			Type:					%s,
			Description:	"%s",
			%s
		},
	`, entry.Parameter, "schema.Type"+entry.Type, entry.Description, extraParams)
	}

	footer := `
	}
)

func GetClusterSchema() map[string]*schema.Schema {
	return clusterSchema
}
	`

	return header + body + footer
}

func getSchemaType(s string) SchemaType {
	if strings.Count(s, "'") == 2 || s == "" {
		return String
	} else if s == "true" || s == "false" {
		return Bool
	} else if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return Array
	}
	return Int
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}