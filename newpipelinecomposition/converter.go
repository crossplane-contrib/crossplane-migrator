package newpipelinecomposition

import (
	"fmt"
	"strings"
	"time"

	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultFunctionRefName = "function-patch-and-transform"

// NewPipelineCompositionFromExisting takes an existing composition and returns a composition
// where the built-in patch & transform has been moved to a function.
// if the existing composition has PipelineMode enabled, it will not change anything
func NewPipelineCompositionFromExisting(c *v1.Composition, functionRefName string) (*v1.Composition, error) {
	if len(c.Spec.Pipeline) > 0 {
		return c, nil
	}

	// prevent null timestamps in the output. k8s apply ignores this field
	if c.ObjectMeta.CreationTimestamp.IsZero() {
		c.ObjectMeta.CreationTimestamp = metav1.NewTime(time.Now())
	}

	cp := &v1.Composition{
		TypeMeta:   c.TypeMeta,
		ObjectMeta: c.ObjectMeta,
		Spec: v1.CompositionSpec{
			CompositeTypeRef:                           c.Spec.CompositeTypeRef,
			WriteConnectionSecretsToNamespace:          c.Spec.DeepCopy().WriteConnectionSecretsToNamespace,
			PublishConnectionDetailsWithStoreConfigRef: c.Spec.PublishConnectionDetailsWithStoreConfigRef.DeepCopy(),
		},
	}
	// Migrate existing input
	input := &Input{
		PatchSets: []v1.PatchSet{},
		Resources: []v1.ComposedTemplate{},
	}
	if c.Spec.Environment != nil {
		cp.Spec.Environment = c.Spec.Environment
	}
	if len(c.Spec.PatchSets) > 0 {
		input.PatchSets = c.Spec.PatchSets
	}
	if len(c.Spec.Resources) > 0 {
		input.Resources = c.Spec.Resources
	}

	// Override function name if provided
	var fr = v1.FunctionReference{Name: defaultFunctionRefName}
	if functionRefName != "" {
		fr.Name = functionRefName
	}

	pipelineMode := v1.CompositionModePipeline

	// Set up the pipeline
	cp.Spec.Mode = &pipelineMode

	ni := NewPatchAndTransformFunctionInput(input)
	cp.Spec.Pipeline = []v1.PipelineStep{
		{
			Step:        "patch-and-transform",
			FunctionRef: fr,
			Input:       ni,
		},
	}
	return cp, nil
}

// Configure Input Function. Since we are migrating legacy Patch & Transform, we convert to
func NewPatchAndTransformFunctionInput(input *Input) *runtime.RawExtension {

	// Populate any missing Fields that are optional in the built-in
	// engine but required in the function
	pi := SetMissingInputFields(input)

	var inputType = map[string]any{
		"apiVersion":  "pt.fn.crossplane.io/v1beta1",
		"kind":        "Resources",
		"environment": pi.Environment,
		"patchSets":   pi.PatchSets,
		"resources":   pi.Resources,
	}

	return &runtime.RawExtension{
		Object: &unstructured.Unstructured{Object: inputType},
	}
}

func SetMissingInputFields(input *Input) *Input {
	var processedInput = &Input{
		Environment: input.Environment,
	}

	processedPatchSet := []v1.PatchSet{}
	for _, patchSet := range input.PatchSets {
		processedPatchSet = append(processedPatchSet, SetMissingPatchSetFields(patchSet))
	}
	processedInput.PatchSets = processedPatchSet

	processedResources := []v1.ComposedTemplate{}
	for idx, resource := range input.Resources {
		processedResources = append(processedResources, SetMissingResourceFields(idx, resource))
	}
	processedInput.Resources = processedResources

	return processedInput
}

func SetMissingPatchSetFields(patchSet v1.PatchSet) v1.PatchSet {
	p := []v1.Patch{}
	for _, patch := range patchSet.Patches {
		p = append(p, SetMissingPatchFields(patch))
	}
	patchSet.Patches = p
	return patchSet
}

func SetMissingPatchFields(patch v1.Patch) v1.Patch {
	if patch.Type == "" {
		patch.Type = v1.PatchTypeFromCompositeFieldPath
	}
	if len(patch.Transforms) == 0 {
		return patch
	}
	t := []v1.Transform{}
	for _, transform := range patch.Transforms {
		t = append(t, SetTransformTypeRequiredFields(transform))
	}
	patch.Transforms = t
	return patch
}

func SetMissingResourceFields(idx int, rs v1.ComposedTemplate) v1.ComposedTemplate {
	if emptyString(rs.Name) {
		kind := rs.Base.Object.GetObjectKind().GroupVersionKind().Kind
		n := strings.ToLower(fmt.Sprintf("%s-%d", kind, idx))
		rs.Name = &n
	}

	cd := []v1.ConnectionDetail{}
	for _, detail := range rs.ConnectionDetails {
		cd = append(cd, SetMissingConnectionDetailFields(detail))
	}
	rs.ConnectionDetails = cd

	patches := []v1.Patch{}
	for _, patch := range rs.Patches {
		patches = append(patches, SetMissingPatchFields(patch))
	}
	rs.Patches = patches
	return rs
}

func emptyString(s *string) bool {
	if s == nil {
		return true
	}

	return *s == ""
}

// This struct is copied from function patch and transform
type Input struct {
	// PatchSets define a named set of patches that may be included by any
	// resource in this Composition. PatchSets cannot themselves refer to other
	// PatchSets.
	//
	// PatchSets are only used by the "Resources" mode of Composition. They
	// are ignored by other modes.
	// +optional
	PatchSets []v1.PatchSet `json:"patchSets,omitempty"`

	// Environment configures the environment in which resources are rendered.
	//
	// THIS IS AN ALPHA FIELD. Do not use it in production. It is not honored
	// unless the relevant Crossplane feature flag is enabled, and may be
	// changed or removed without notice.
	// +optional
	Environment *v1.EnvironmentConfiguration `json:"environment,omitempty"`

	// Resources is a list of resource templates that will be used when a
	// composite resource referring to this composition is created.
	//
	// Resources are only used by the "Resources" mode of Composition. They are
	// ignored by other modes.
	// +optional
	Resources []v1.ComposedTemplate `json:"resources,omitempty"`
}

// SetTransformTypeRequiredFields sets fields that are required with
// function-patch-and-transform but were optional with the built-in engine
func SetTransformTypeRequiredFields(tt v1.Transform) v1.Transform {
	if tt.Type == "" {
		if tt.Math != nil {
			tt.Type = v1.TransformTypeMath
		}
		if tt.String != nil {
			tt.Type = v1.TransformTypeString
		}
	}
	if tt.Type == v1.TransformTypeMath && tt.Math.Type == "" {
		if tt.Math.ClampMin != nil {
			tt.Math.Type = v1.MathTransformTypeClampMin
		}
		if tt.Math.ClampMax != nil {
			tt.Math.Type = v1.MathTransformTypeClampMax
		}
		if tt.Math.Multiply != nil {
			tt.Math.Type = v1.MathTransformTypeMultiply
		}
	}

	if tt.Type == v1.TransformTypeString && tt.String.Type == "" {
		if tt.String.Format != nil {
			tt.String.Type = v1.StringTransformTypeFormat
		}
		if tt.String.Convert != nil {
			tt.String.Type = v1.StringTransformTypeConvert
		}
		if tt.String.Regexp != nil {
			tt.String.Type = v1.StringTransformTypeRegexp
		}
	}
	return tt
}

func SetMissingConnectionDetailFields(sk v1.ConnectionDetail) v1.ConnectionDetail {
	fv := v1.ConnectionDetailTypeFromValue
	ffp := v1.ConnectionDetailTypeFromFieldPath
	fcsk := v1.ConnectionDetailTypeFromConnectionSecretKey

	// Only one of the values should be set, but we are not validating it here
	nsk := v1.ConnectionDetail{
		Name:                    sk.Name,
		Value:                   sk.Value,
		FromConnectionSecretKey: sk.FromConnectionSecretKey,
		FromFieldPath:           sk.FromFieldPath,
	}
	// Type is now required
	if nsk.Type == nil {
		if sk.Value != nil {
			nsk.Type = &fv
		}
		if sk.FromFieldPath != nil {
			nsk.Type = &ffp
		}
		if sk.FromConnectionSecretKey != nil {
			nsk.Type = &fcsk
		}
	}
	// Name is also required
	if nsk.Name == nil {
		switch t := nsk.Type; t {
		case &fcsk:
			nsk.Name = sk.FromConnectionSecretKey
		}
		// FromValue and FromFieldPath should have a name, skip implementation for now
	}
	return nsk
}
